package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/log.go/log"
	"github.com/pentops/outbox.pg.go/outbox"
	"github.com/pentops/outbox.pg.go/outboxtest"
	"github.com/pentops/pgtest.go/pgtest"
	"github.com/pentops/registry/anyfs"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/registry/integration/mocks"
	"github.com/pentops/registry/japi"
	"github.com/pentops/registry/packagestore"
	"github.com/pentops/registry/service"
	"github.com/pentops/registry/state"
	"github.com/rs/cors"
)

type Universe struct {
	Outbox        *outboxtest.OutboxAsserter
	GithubCommand github_spb.GithubCommandServiceClient
	RepoQuery     github_spb.RepoQueryServiceClient
	WebhookTopic  github_pb.WebhookTopicClient

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
		uu.Outbox.AssertNoMessages(t)
		return nil
	})

	return stepper, uu
}

const TestVersion = "test-version"

// setupUniverse should only be called from the Setup callback, it is effectively
// a method but shouldn't show up there.
func setupUniverse(ctx context.Context, t flowtest.Asserter, uu *Universe) {
	t.Helper()

	conn := pgtest.GetTestDB(t, pgtest.WithDir("../ext/db"))

	uu.Outbox = outboxtest.NewOutboxAsserter(t, conn)
	uu.Github = mocks.NewGithubMock()

	grpcPair := flowtest.NewGRPCPair(t, service.GRPCMiddleware()...)

	outboxPub, err := outbox.NewDBPublisher(conn)
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
	github_pb.RegisterWebhookTopicServer(grpcPair.Server, webhookWorker)
	uu.WebhookTopic = github_pb.NewWebhookTopicClient(grpcPair.Client)

	commandService, err := service.NewRepoCommandService(conn, states)
	if err != nil {
		t.Fatalf("failed to create github command service: %v", err)
	}
	github_spb.RegisterGithubCommandServiceServer(grpcPair.Server, commandService)
	uu.GithubCommand = github_spb.NewGithubCommandServiceClient(grpcPair.Client)

	queryService, err := service.NewRepoQueryService(conn, states)
	if err != nil {
		t.Fatalf("failed to create github query service: %v", err)
	}
	github_spb.RegisterRepoQueryServiceServer(grpcPair.Server, queryService)
	uu.RepoQuery = github_spb.NewRepoQueryServiceClient(grpcPair.Client)

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