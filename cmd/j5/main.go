package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bufbuild/protoyaml-go"
	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-go/github/v1/github_pb"
	"github.com/pentops/registry/anyfs"
	"github.com/pentops/registry/builder"
	"github.com/pentops/registry/github"
	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/registry/japi"
	"github.com/pentops/registry/messaging"
	"github.com/pentops/runner"
	"github.com/pentops/runner/commander"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.daemonl.com/log/grpc_log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	gomodGroup := commander.NewCommandSet()
	gomodGroup.Add("upload", commander.NewCommand(gomodproxy.RunUploadCommand))
	cmdGroup.Add("gomod", gomodGroup)

	japiGroup := commander.NewCommandSet()
	japiGroup.Add("push", commander.NewCommand(runPushAPI))
	cmdGroup.Add("japi", japiGroup)

	protoGroup := commander.NewCommandSet()
	protoGroup.Add("build", commander.NewCommand(runProtoBuild))
	cmdGroup.Add("proto", protoGroup)

	cmdGroup.Add("serve", commander.NewCommand(runCombinedServer))

	cmdGroup.Add("test", commander.NewCommand(runTestBuild))

	cmdGroup.RunMain("registry", Version)
}

func runTestBuild(ctx context.Context, cfg struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
}) error {
	remote := builder.NewNOPUploader()

	japiConfigData, err := os.ReadFile(filepath.Join(cfg.Source, "jsonapi.yaml"))
	if err != nil {
		return err
	}
	japiConfig := &config_j5pb.Config{}
	if err := protoyaml.Unmarshal(japiConfigData, japiConfig); err != nil {
		return err
	}

	commitInfo := &builder_j5pb.CommitInfo{
		Hash:    "test",
		Owner:   "test",
		Repo:    "test",
		Time:    timestamppb.New(time.Now()),
		Aliases: []string{},
	}

	dockerWrapper, err := builder.NewDockerWrapper(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	bb := builder.NewBuilder(dockerWrapper, remote)

	return bb.BuildAll(ctx, japiConfig, cfg.Source, commitInfo)
}

func runProtoBuild(ctx context.Context, cfg struct {
	Source        string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	PackagePrefix string `flag:"package-prefix" env:"PACKAGE_PREFIX" default:""`
	Storage       string `env:"REGISTRY_STORAGE"`

	CommitHash    string   `flag:"commit-hash" env:"COMMIT_HASH" default:""`
	CommitTime    string   `flag:"commit-time" env:"COMMIT_TIME" default:""`
	CommitAliases []string `flag:"commit-alias" env:"COMMIT_ALIAS" default:""`

	GitAuto bool `flag:"git-auto" env:"COMMIT_INFO_GIT_AUTO" default:"false" description:"Automatically pull commit info from git"`
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	s3fs, err := anyfs.NewS3FS(s3Client, cfg.Storage)
	if err != nil {
		return err
	}

	remote := builder.NewFSUploader(s3fs)

	japiConfigData, err := os.ReadFile(filepath.Join(cfg.Source, "jsonapi.yaml"))
	if err != nil {
		return err
	}
	japiConfig := &config_j5pb.Config{}
	if err := protoyaml.Unmarshal(japiConfigData, japiConfig); err != nil {
		return err
	}

	var commitInfo *builder_j5pb.CommitInfo
	if cfg.GitAuto {
		commitInfo, err = builder.ExtractGitMetadata(ctx, japiConfig.Git, cfg.Source)
		if err != nil {
			return err
		}
	} else if cfg.CommitHash == "" || cfg.CommitTime == "" {
		return fmt.Errorf("commit hash and time are required, or set --git-auto")
	} else {
		commitInfo.Hash = cfg.CommitHash
		commitTime, err := time.Parse(time.RFC3339, cfg.CommitTime)
		if err != nil {
			return fmt.Errorf("parsing commit time: %w", err)
		}
		commitInfo.Time = timestamppb.New(commitTime)
		commitInfo.Aliases = cfg.CommitAliases
	}

	dockerWrapper, err := builder.NewDockerWrapper(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	bb := builder.NewBuilder(dockerWrapper, remote)

	return bb.BuildAll(ctx, japiConfig, cfg.Source, commitInfo)
}

func runPushAPI(ctx context.Context, cfg struct {
	Source  string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Storage string `env:"REGISTRY_STORAGE"`

	CommitHash    string   `flag:"commit-hash" env:"COMMIT_HASH" default:""`
	CommitTime    string   `flag:"commit-time" env:"COMMIT_TIME" default:""`
	CommitAliases []string `flag:"commit-alias" env:"COMMIT_ALIAS" default:""`

	GitAuto bool `flag:"git-auto" env:"COMMIT_INFO_GIT_AUTO" default:"false" description:"Automatically pull commit info from git"`
}) error {

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	s3fs, err := anyfs.NewS3FS(s3Client, cfg.Storage)
	if err != nil {
		return err
	}

	remote := builder.NewFSUploader(s3fs)

	//return remote.UploadJsonAPI(ctx, cfg.Version, bb)
	_ = remote
	_ = bb
	return nil

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
	HTTPPort    int      `env:"HTTP_PORT" default:"8081"`
	GRPCPort    int      `env:"GRPC_PORT" default:"8080"`
	Storage     string   `env:"REGISTRY_STORAGE"`
	SourceRepos []string `env:"SOURCE_REPOS"`
	SNSPrefix   string   `env:"SNS_PREFIX"`
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

	remote := builder.NewFSUploader(s3fs)

	dockerWrapper, err := builder.NewDockerWrapper(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	githubClient, err := github.NewEnvClient(ctx)
	if err != nil {
		return err
	}

	builderHandler := builder.NewBuildWorker(builder.NewBuilder(dockerWrapper, remote), githubClient)

	var publisher github.Publisher
	if strings.ToLower(cfg.SNSPrefix) == "local" {
		publisher = localPublisher{
			handler: builderHandler,
		}
	} else {
		publisher = messaging.NewSNSPublisher(snsClient, cfg.SNSPrefix)
	}

	githubWorker, err := github.NewWebhookWorker(githubClient, publisher, cfg.SourceRepos)
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
		builder_j5pb.RegisterBuilderTopicServer(grpcServer, builderHandler)
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

type localPublisher struct {
	handler builder_j5pb.BuilderTopicServer
}

func (l localPublisher) Publish(ctx context.Context, msg messaging.Message) error {

	switch msg := msg.(type) {
	case *builder_j5pb.BuildProtoMessage:
		_, err := l.handler.BuildProto(ctx, msg)
		return err
	case *builder_j5pb.BuildAPIMessage:
		_, err := l.handler.BuildAPI(ctx, msg)
		return err
	default:
		return fmt.Errorf("unknown local publisher message: %T", msg)
	}

}
