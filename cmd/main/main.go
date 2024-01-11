package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bufbuild/protoyaml-go"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/anyfs"
	"github.com/pentops/registry/builder"
	"github.com/pentops/registry/github"
	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/registry/japi"
	"github.com/pentops/runner/commander"
	"github.com/rs/cors"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	cmdGroup.RunMain("registry", Version)
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
	japiConfig := &jsonapi_pb.Config{}
	if err := protoyaml.Unmarshal(japiConfigData, japiConfig); err != nil {
		return err
	}

	var commitInfo *builder.CommitInfo
	if cfg.GitAuto {
		commitInfo, err = builder.ExtractGitMetadata(ctx, japiConfig.Git, cfg.Source)
		if err != nil {
			return err
		}
	} else if cfg.CommitHash == "" || cfg.CommitTime == "" {
		return fmt.Errorf("commit hash and time are required, or set --git-auto")
	} else {
		commitInfo.Hash = cfg.CommitHash
		commitInfo.Time, err = time.Parse(time.RFC3339, cfg.CommitTime)
		if err != nil {
			return fmt.Errorf("parsing commit time: %w", err)
		}
		commitInfo.Aliases = cfg.CommitAliases
	}

	dockerWrapper, err := builder.NewDockerWrapper(japiConfig.DockerRegistryAuths)
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

func TriggerHandler(githubClient *github.Client, uploader builder.IUploader) http.Handler {

	fetchAndBuild := func(ctx context.Context, owner string, repo string, version string) error {

		workDir, err := os.MkdirTemp("", "trigger")
		if err != nil {
			return err
		}

		defer os.RemoveAll(workDir)

		commitInfo, err := githubClient.GetCommit(ctx, owner, repo, version)
		if err != nil {
			return err
		}
		if err := githubClient.GetContent(ctx, owner, repo, commitInfo.Hash, workDir); err != nil {
			return err
		}

		japiConfigData, err := os.ReadFile(filepath.Join(workDir, "jsonapi.yaml"))
		if err != nil {
			return err
		}
		japiConfig := &jsonapi_pb.Config{}
		if err := protoyaml.Unmarshal(japiConfigData, japiConfig); err != nil {
			return err
		}

		dockerWrapper, err := builder.NewDockerWrapper(japiConfig.DockerRegistryAuths)
		if err != nil {
			return err
		}

		bb := builder.NewBuilder(dockerWrapper, uploader)

		return bb.BuildAll(ctx, japiConfig, workDir, commitInfo)
	}

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

		if err := fetchAndBuild(r.Context(), owner, repo, version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
}

func runCombinedServer(ctx context.Context, cfg struct {
	Port    int    `env:"PORT" default:"8080"`
	Storage string `env:"REGISTRY_STORAGE"`
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

	japiHandler, err := japi.NewRegistry(ctx, s3fs.SubFS("japi"))
	if err != nil {
		return err
	}

	gomodData, err := gomodproxy.NewS3PackageSrc(ctx, s3fs.SubFS("gomod"))
	if err != nil {
		return err

	}

	remote := builder.NewFSUploader(s3fs)

	githubClient, err := github.NewEnvClient(ctx)
	if err != nil {
		return err
	}

	genericCORS := cors.Default()
	mux := http.NewServeMux()
	mux.Handle("/registry/v1/", genericCORS.Handler(http.StripPrefix("/registry/v1", japiHandler)))
	mux.Handle("/gopkg/", http.StripPrefix("/gopkg", gomodproxy.Handler(gomodData)))
	mux.Handle("/trigger/v1/", http.StripPrefix("/trigger/v1", TriggerHandler(githubClient, remote)))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	log.WithField(ctx, "port", cfg.Port).Info("Begin Registry Server")

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx) // nolint:errcheck
	}()

	return httpServer.ListenAndServe()
}
