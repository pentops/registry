package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/pentops/j5/builder"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/internal/anyfs"
	"github.com/pentops/registry/internal/buildwrap"
	"github.com/pentops/registry/internal/github"
	"github.com/pentops/registry/internal/gomodproxy"
	"github.com/pentops/registry/internal/packagestore"
	"github.com/pentops/registry/internal/service"
	"github.com/pentops/registry/internal/state"
	"github.com/pentops/runner"
	"github.com/pentops/runner/commander"
	"github.com/pressly/goose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/o5-messaging/outbox"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("serve", commander.NewCommand(runCombinedServer))
	cmdGroup.Add("migrate", commander.NewCommand(runMigrate))

	cmdGroup.RunMain("registry", Version)
}

func runMigrate(ctx context.Context, cfg struct {
	MigrationsDir string `env:"MIGRATIONS_DIR" default:"./ext/db"`
	service.DBConfig
}) error {

	db, err := cfg.OpenDatabase(ctx)
	if err != nil {
		return err
	}

	return goose.Up(db, "/migrations")
}

func runCombinedServer(ctx context.Context, cfg struct {
	HTTPPort int    `env:"HTTP_PORT" default:"8081"`
	GRPCPort int    `env:"GRPC_PORT" default:"8080"`
	Storage  string `env:"REGISTRY_STORAGE"`
	service.DBConfig
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := cfg.DBConfig.OpenDatabase(ctx)
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(awsConfig)

	s3fs, err := anyfs.NewS3FS(s3Client, cfg.Storage)
	if err != nil {
		return err
	}

	pkgStore, err := packagestore.NewPackageStore(db, s3fs)
	if err != nil {
		return err
	}

	dockerWrapper, err := builder.NewRunner(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	githubClient, err := github.NewEnvClient(ctx)
	if err != nil {
		return err
	}

	dbPublisher, err := outbox.NewDirectPublisher(db, outbox.DefaultSender)
	if err != nil {
		return err
	}

	j5Builder := builder.NewBuilder(dockerWrapper)

	buildWorker := buildwrap.NewBuildWorker(j5Builder, githubClient, pkgStore, dbPublisher)

	refStore, err := service.NewRefStore(db)
	if err != nil {
		return err
	}

	githubWorker, err := service.NewWebhookWorker(refStore, githubClient, dbPublisher)
	if err != nil {
		return err
	}

	stateMachines, err := state.NewStateMachines()
	if err != nil {
		return err
	}

	githubCommand, err := service.NewGithubCommandService(db, stateMachines, githubWorker)
	if err != nil {
		return err
	}

	githubQuery, err := service.NewGithubQueryService(db, stateMachines)
	if err != nil {
		return err
	}

	registryDownloadService := service.NewRegistryService(pkgStore)

	runGroup := runner.NewGroup(runner.WithName("main"), runner.WithCancelOnSignals())

	runGroup.Add("httpServer", func(ctx context.Context) error {
		handler := gomodproxy.Handler(pkgStore)

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: handler,
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
			service.GRPCMiddleware()...,
		))

		githubWorker.RegisterGRPC(grpcServer)
		buildWorker.RegisterGRPC(grpcServer)
		githubCommand.RegisterGRPC(grpcServer)
		githubQuery.RegisterGRPC(grpcServer)
		registryDownloadService.RegisterGRPC(grpcServer)

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
