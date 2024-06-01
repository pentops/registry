package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/pentops/jsonapi/builder/builder"
	"github.com/pentops/jsonapi/builder/docker"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/anyfs"
	"github.com/pentops/registry/buildwrap"
	"github.com/pentops/registry/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/github"
	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/registry/japi"
	"github.com/pentops/registry/messaging"
	"github.com/pentops/runner"
	"github.com/pentops/runner/commander"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.daemonl.com/log/grpc_log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("serve", commander.NewCommand(runCombinedServer))

	cmdGroup.RunMain("registry", Version)
}

func TriggerHandler(githubWorker github_pb.WebhookTopicServer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /$owner/$repo/$version
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 4 {
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}

		owner := parts[1]
		repo := parts[2]
		version := parts[3]

		_, err := githubWorker.Push(r.Context(), &github_pb.PushMessage{
			Owner: owner,
			Repo:  repo,
			Ref:   "ignored",
			After: version,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func runCombinedServer(ctx context.Context, cfg struct {
	HTTPPort         int      `env:"HTTP_PORT" default:"8081"`
	GRPCPort         int      `env:"GRPC_PORT" default:"8080"`
	Storage          string   `env:"REGISTRY_STORAGE"`
	SourceRepos      []string `env:"SOURCE_REPOS"`
	SourceCheckRepos []string `env:"SOURCE_CHECK_REPOS"`
	SNSPrefix        string   `env:"SNS_PREFIX"`
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)
	snsClient := sns.NewFromConfig(awsConfig)

	s3fs, err := anyfs.NewS3FS(s3Client, cfg.Storage)
	if err != nil {
		return err
	}

	japiHandler, err := japi.NewRegistry(ctx, s3fs.SubFS("japi"))
	if err != nil {
		return err
	}

	gomodData, err := gomodproxy.NewS3PackageSrc(ctx, s3fs.SubFS("gomod"))
	if err != nil {
		return err
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	githubClient, err := github.NewEnvClient(ctx)
	if err != nil {
		return err
	}

	j5Builder := builder.NewBuilder(dockerWrapper)
	buildWorker := buildwrap.NewBuildWorker(j5Builder, githubClient, s3fs)
	publisher := messaging.NewSNSPublisher(snsClient, cfg.SNSPrefix)

	githubWorker, err := github.NewWebhookWorker(githubClient, publisher, cfg.SourceRepos, cfg.SourceCheckRepos)
	if err != nil {
		return err
	}

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {

		genericCORS := cors.Default()
		mux := http.NewServeMux()
		mux.Handle("/registry/v1/", genericCORS.Handler(http.StripPrefix("/registry/v1", japiHandler)))
		mux.Handle("/gopkg/", http.StripPrefix("/gopkg", gomodproxy.Handler(gomodData)))
		mux.Handle("/trigger/v1/", http.StripPrefix("/trigger/v1", TriggerHandler(githubWorker)))

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
		}
		log.WithField(ctx, "port", cfg.HTTPPort).Info("Begin Registry Server")

		go func() {
			<-ctx.Done()
			httpServer.Shutdown(ctx) // nolint:errcheck
		}()

		return httpServer.ListenAndServe()
	})

	runGroup.Add("grpcServer", func(ctx context.Context) error {
		grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
			grpc_log.UnaryServerInterceptor(log.DefaultContext, log.DefaultTrace, log.DefaultLogger),
		))
		github_pb.RegisterWebhookTopicServer(grpcServer, githubWorker)
		builder_tpb.RegisterBuilderTopicServer(grpcServer, buildWorker)
		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			return err
		}
		log.WithField(ctx, "port", cfg.GRPCPort).Info("Begin Worker Server")
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop() // nolint:errcheck
		}()

		return grpcServer.Serve(lis)
	})

	return runGroup.Run(ctx)
}