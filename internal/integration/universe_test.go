package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-messaging/outbox"
	"github.com/pentops/o5-messaging/outbox/outboxtest"
	"github.com/pentops/pgtest.go/pgtest"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/internal/anyfs"
	"github.com/pentops/registry/internal/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/internal/gomodproxy"
	"github.com/pentops/registry/internal/integration/mocks"
	"github.com/pentops/registry/internal/japi"
	"github.com/pentops/registry/internal/packagestore"
	"github.com/pentops/registry/internal/service"
	"github.com/pentops/registry/internal/state"
	"github.com/rs/cors"
)

type Universe struct {
	Outbox        *outboxtest.OutboxAsserter
	GithubCommand github_spb.GithubCommandServiceClient
	GithubQuery   github_spb.GithubQueryServiceClient
	WebhookTopic  github_tpb.WebhookTopicClient
	BuilderReply  builder_tpb.BuilderReplyTopicClient

	PackageStore *packagestore.PackageStore

	Github *mocks.GithubMock

	HTTPHandler http.Handler
}

func NewUniverse(t *testing.T) (*flowtest.Stepper[*testing.T], *Universe) {
	name := t.Name()
	stepper := flowtest.NewStepper[*testing.T](name)
	uu := &Universe{}

	stepper.Setup(func(ctx context.Context, t flowtest.Asserter) error {
		log.DefaultLogger = log.NewCallbackLogger(stepper.LevelLog)
		setupUniverse(ctx, t, uu)
		return nil
	})

	stepper.PostStepHook(func(ctx context.Context, t flowtest.Asserter) error {
		uu.Outbox.AssertEmpty(t)
		return nil
	})

	return stepper, uu
}

const TestVersion = "test-version"

// setupUniverse should only be called from the Setup callback, it is effectively
// a method but shouldn't show up there.
func setupUniverse(ctx context.Context, t flowtest.Asserter, uu *Universe) {
	t.Helper()

	conn := pgtest.GetTestDB(t, pgtest.WithDir("../../ext/db"))

	uu.Outbox = outboxtest.NewOutboxAsserter(t, conn)
	uu.Github = mocks.NewGithubMock()

	grpcPair := flowtest.NewGRPCPair(t, service.GRPCMiddleware()...)

	outboxPub, err := outbox.NewDirectPublisher(conn, outbox.DefaultSender)
	if err != nil {
		t.Fatalf("failed to create outbox publisher: %v", err)
	}

	states, err := state.NewStateMachines()
	if err != nil {
		t.Fatalf("failed to create state machines: %v", err)
	}

	refs, err := service.NewRefStore(conn)
	if err != nil {
		t.Fatalf("failed to create ref store: %v", err)
	}

	tmpfs, err := anyfs.NewTempFS(ctx)
	if err != nil {
		t.Fatalf("failed to create temp fs: %v", err)
	}

	pkgStore, err := packagestore.NewPackageStore(conn, tmpfs)
	if err != nil {
		t.Fatalf("failed to create package store: %v", err)
	}

	uu.PackageStore = pkgStore

	{
		genericCORS := cors.Default()
		mux := http.NewServeMux()
		mux.Handle("/registry/v1/", genericCORS.Handler(http.StripPrefix("/registry/v1", japi.Handler(pkgStore))))
		mux.Handle("/gopkg/", http.StripPrefix("/gopkg", gomodproxy.Handler(pkgStore)))
		uu.HTTPHandler = mux
	}

	webhookWorker, err := service.NewWebhookWorker(refs, uu.Github, outboxPub)
	if err != nil {
		t.Fatalf("failed to create webhook worker: %v", err)
	}
	github_tpb.RegisterWebhookTopicServer(grpcPair.Server, webhookWorker)
	uu.WebhookTopic = github_tpb.NewWebhookTopicClient(grpcPair.Client)

	commandService, err := service.NewGithubCommandService(conn, states)
	if err != nil {
		t.Fatalf("failed to create github command service: %v", err)
	}
	github_spb.RegisterGithubCommandServiceServer(grpcPair.Server, commandService)
	uu.GithubCommand = github_spb.NewGithubCommandServiceClient(grpcPair.Client)

	queryService, err := service.NewGithubQueryService(conn, states)
	if err != nil {
		t.Fatalf("failed to create github query service: %v", err)
	}
	github_spb.RegisterGithubQueryServiceServer(grpcPair.Server, queryService)
	uu.GithubQuery = github_spb.NewGithubQueryServiceClient(grpcPair.Client)

	builder_tpb.RegisterBuilderReplyTopicServer(grpcPair.Server, webhookWorker)
	uu.BuilderReply = builder_tpb.NewBuilderReplyTopicClient(grpcPair.Client)

	grpcPair.ServeUntilDone(t, ctx)
}

type HTTPResponse struct {
	Body       []byte
	StatusCode int
}

func (uu *Universe) HTTPGet(ctx context.Context, path string) HTTPResponse {
	req := httptest.NewRequest("GET", path, nil)
	req = req.WithContext(ctx)

	res := httptest.NewRecorder()
	uu.HTTPHandler.ServeHTTP(res, req)

	out := HTTPResponse{
		Body:       res.Body.Bytes(),
		StatusCode: res.Code,
	}

	log.WithFields(ctx, map[string]interface{}{
		"status": res.Code,
		"body":   string(out.Body),
		"path":   path,
	}).Info("HTTP GET")

	return out
}
