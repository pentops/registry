package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/builder"
	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/registry/japi"
	"github.com/pentops/runner/commander"
	"github.com/rs/cors"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"gopkg.in/yaml.v3"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	gomodGroup := commander.NewCommandSet()
	gomodGroup.Add("serve", commander.NewCommand(runGomodServer))
	gomodGroup.Add("upload", commander.NewCommand(gomodproxy.RunUploadCommand))
	cmdGroup.Add("gomod", gomodGroup)

	japiGroup := commander.NewCommandSet()
	japiGroup.Add("serve", commander.NewCommand(runJapiRegistry))
	japiGroup.Add("push", commander.NewCommand(runPushAPI))
	cmdGroup.Add("japi", japiGroup)

	protoGroup := commander.NewCommandSet()
	protoGroup.Add("build", commander.NewCommand(runProtoBuild))
	protoGroup.Add("request", commander.NewCommand(runProtoBuildRequest))
	cmdGroup.Add("proto", protoGroup)

	cmdGroup.Add("serve", commander.NewCommand(runCombinedServer))

	cmdGroup.RunMain("registry", Version)
}

func runProtoBuildRequest(ctx context.Context, cfg struct {
	Source        string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Parameter     string `flag:"parameter" default:""`
	PackagePrefix string `flag:"package-prefix"`
}) error {
	protoSource, err := structure.ReadFileDescriptorSet(ctx, cfg.Source)
	if err != nil {
		return err
	}

	input, err := structure.CodeGeneratorRequestFromDescriptors(structure.CodeGenOptions{
		PackagePrefix: cfg.PackagePrefix,
	}, protoSource)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(input)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(bb)
	return err

}

func runProtoBuild(ctx context.Context, cfg struct {
	Source        string `flag:"src" default:"." env:"PROTO_SOURCE" description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	PackagePrefix string `flag:"package-prefix" env:"PACKAGE_PREFIX" default:""`
	GoModFile     string `flag:"gomod-file" env:"GOMOD_FILE" default:"go.mod"`

	CommitHash    string   `flag:"commit-hash" env:"COMMIT_HASH" default:""`
	CommitTime    string   `flag:"commit-time" env:"COMMIT_TIME" default:""`
	CommitAliases []string `flag:"commit-alias" env:"COMMIT_ALIAS" default:""`
	GitAuto       bool     `flag:"git-auto" env:"COMMIT_INFO_GIT_AUTO" default:"false" description:"Automatically pull commit info from git"`

	GomodRemote string `env:"GOMOD_REMOTE"`
	BufGenFile  string `flag:"buf-gen-file" env:"BUF_GEN_FILE" default:"buf.gen.yaml"`
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	remote, err := gomodproxy.NewS3PackageSrc(ctx, s3Client, cfg.GomodRemote)
	if err != nil {
		return err
	}

	generateSource := cfg.BufGenFile
	if !filepath.IsAbs(generateSource) {
		generateSource = filepath.Join(cfg.Source, generateSource)
	}
	generateBytes, err := os.ReadFile(generateSource)
	if err != nil {
		return err
	}

	generateSpec := &builder.BufGenerateConfig{}

	if strings.HasSuffix(generateSource, ".json") {
		if err := json.Unmarshal(generateBytes, &generateSpec); err != nil {
			return err
		}
	} else if strings.HasSuffix(generateSource, ".yaml") || strings.HasSuffix(generateSource, ".yml") {
		if err := yaml.Unmarshal(generateBytes, &generateSpec); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown generate file type")
	}

	builders, err := builder.ConvertBufGenerateSpec(generateSpec)
	if err != nil {
		return err
	}

	var commitTime time.Time
	var commitHash string
	var commitAliases []string

	if cfg.GitAuto {
		repo, err := git.PlainOpen(cfg.Source)
		if err != nil {
			return err
		}

		head, err := repo.Head()
		if err != nil {
			return err
		}

		commit, err := repo.CommitObject(head.Hash())
		if err != nil {
			return err
		}

		commitTime = commit.Committer.When
		commitHash = commit.Hash.String()

		commitAliases = append(commitAliases, commitHash)

		headName := head.Name()
		if headName.IsBranch() {
			commitAliases = append(commitAliases, headName.Short())
			commitAliases = append(commitAliases, string(headName))

			// TODO: Make this configurable
			if headName == "refs/heads/main" {
				commitAliases = append(commitAliases, "latest")
			}
		}

		// TODO: Tags

		log.WithFields(ctx, map[string]interface{}{
			"commitHash":    commitHash,
			"commitTime":    commitTime,
			"commitAliases": commitAliases,
		}).Info("Resolved Git Commit Info")
	} else if cfg.CommitHash == "" || cfg.CommitTime == "" {
		return fmt.Errorf("commit hash and time are required, or set --git-auto")
	} else {
		commitHash = cfg.CommitHash
		commitTime, err = time.Parse(time.RFC3339, cfg.CommitTime)
		if err != nil {
			return fmt.Errorf("parsing commit time: %w", err)
		}
		commitAliases = cfg.CommitAliases
	}

	protoSource, err := structure.ReadFileDescriptorSet(ctx, cfg.Source)
	if err != nil {
		return err
	}

	input, err := structure.CodeGeneratorRequestFromDescriptors(structure.CodeGenOptions{
		PackagePrefix: cfg.PackagePrefix,
	}, protoSource)
	if err != nil {
		return err
	}

	goModFileSource := cfg.GoModFile
	if !filepath.IsAbs(goModFileSource) {
		goModFileSource = filepath.Join(cfg.Source, goModFileSource)
	}
	gomodData, err := os.ReadFile(goModFileSource)
	if err != nil {
		return err
	}

	return builder.BuildImage(ctx, builder.BuildSpec{
		GoModFile:     gomodData,
		CommitTime:    commitTime,
		CommitHash:    commitHash,
		Builders:      builders,
		CommitAliases: commitAliases,
	}, input, remote)
}

func runPushAPI(ctx context.Context, cfg struct {
	Source  string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Version string `flag:"version" default:"" description:"Version to push"`
	Latest  bool   `flag:"latest" description:"Push as latest"`
	Bucket  string `flag:"bucket" description:"S3 bucket to push to"`
	Prefix  string `flag:"prefix" description:"S3 prefix to push to"`
}) error {

	if (!cfg.Latest) && cfg.Version == "" {
		return fmt.Errorf("version, latest or both are required")
	}

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	versions := []string{}

	if cfg.Latest {
		versions = append(versions, "latest")
	}

	if cfg.Version != "" {
		versions = append(versions, cfg.Version)
	}

	destinations := make([]string, len(versions))
	for i, version := range versions {
		p := path.Join(cfg.Prefix, image.Registry.Organization, image.Registry.Name, version, "image.bin")
		destinations[i] = fmt.Sprintf("s3://%s/%s", cfg.Bucket, p)
	}

	return pushS3(ctx, bb, destinations...)
}

func pushS3(ctx context.Context, bb []byte, destinations ...string) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)
	for _, dest := range destinations {
		s3URL, err := url.Parse(dest)
		if err != nil {
			return err
		}
		if s3URL.Scheme != "s3" || s3URL.Host == "" {
			return fmt.Errorf("invalid s3 url: %s", dest)
		}

		log.WithField(ctx, "dest", dest).Info("Uploading to S3")

		// url.Parse will take s3://foobucket/keyname and turn keyname into "/keyname" which we want to be "keyname"
		k := strings.Replace(s3URL.Path, "/", "", 1)

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &s3URL.Host,
			Key:    &k,
			Body:   strings.NewReader(string(bb)),
		})

		if err != nil {
			return fmt.Errorf("failed to upload object: %w", err)
		}
	}

	return nil
}

func runCombinedServer(ctx context.Context, cfg struct {
	Port        int    `env:"PORT" default:"8080"`
	GomodRemote string `env:"GOMOD_REMOTE"`
	JAPIRemote  string `env:"JAPI_REMOTE"`
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	japiHandler, err := japi.NewRegistry(ctx, s3Client, cfg.JAPIRemote)
	if err != nil {
		return err
	}

	gomodData, err := gomodproxy.NewS3PackageSrc(ctx, s3Client, cfg.GomodRemote)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/registry/v1/", http.StripPrefix("/registry/v1", japiHandler))
	mux.Handle("/gopkg/", http.StripPrefix("/gopkg", gomodproxy.Handler(gomodData)))

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

func runGomodServer(ctx context.Context, cfg struct {
	Port int    `env:"PORT" default:"8080"`
	Src  string `env:"GOMOD_REMOTE"`
}) error {

	var data gomodproxy.ModProvider
	var err error

	if strings.HasPrefix(cfg.Src, "s3://") {
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		s3Client := s3.NewFromConfig(awsConfig)

		data, err = gomodproxy.NewS3PackageSrc(ctx, s3Client, cfg.Src)
		if err != nil {
			return err
		}
	} else {
		root := os.DirFS(cfg.Src)
		data, err = gomodproxy.NewLocalPackageMap(ctx, root)
		if err != nil {
			return err
		}
	}

	return gomodproxy.Serve(ctx, cfg.Port, data)
}

func runJapiRegistry(ctx context.Context, cfg struct {
	RegistryPort   int    `env:"REGISTRY_PORT" default:""`
	RegistryBucket string `env:"JAPI_REMOTE" default:""`
}) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	handler, err := japi.NewRegistry(ctx, s3Client, cfg.RegistryBucket)
	if err != nil {
		return err
	}

	httpHandler := http.StripPrefix("/registry/v1", handler)
	httpHandler = cors.Default().Handler(httpHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.RegistryPort),
		Handler: httpHandler,
	}
	log.WithField(ctx, "port", cfg.RegistryPort).Info("Begin Registry Server")

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx) // nolint:errcheck
	}()

	return httpServer.ListenAndServe()

}
