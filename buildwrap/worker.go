package buildwrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pentops/jsonapi/builder/builder"
	"github.com/pentops/jsonapi/builder/git"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/schema/structure"
	"github.com/pentops/jsonapi/schema/swagger"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/github"
	"github.com/pentops/registry/gomodproxy"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BuildWorker struct {
	builder_tpb.UnimplementedBuilderTopicServer

	builder J5Builder
	github  IGithub

	goModUploader RemoteWithMetadata
	j5Uploader    RemoteWithMetadata
}

type J5Builder interface {
	BuildProto(ctx context.Context, source builder.Source, dest builder.FS, builderName string, logWriter io.Writer) error
}

type IGithub interface {
	GetContent(ctx context.Context, ref github.RepoRef, intoDir string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*source_j5pb.CommitInfo, error)
	UpdateCheckRun(ctx context.Context, ref github.RepoRef, checkRun *builder_tpb.CheckRun, status github.CheckRunUpdate) error
}

func NewBuildWorker(builder J5Builder, github IGithub, rootUploader RemoteWithMetadata) *BuildWorker {

	goModUploader := SubRemote(rootUploader, "gomod")
	j5Uploader := SubRemote(rootUploader, "japi")

	return &BuildWorker{
		builder:       builder,
		github:        github,
		goModUploader: goModUploader,
		j5Uploader:    j5Uploader,
	}
}

func (bw *BuildWorker) updateCheckRun(ctx context.Context, commit *source_j5pb.CommitInfo, checkRun *builder_tpb.CheckRun, status github.CheckRunUpdate) error {
	if checkRun == nil {
		return nil
	}

	if err := bw.github.UpdateCheckRun(ctx, github.RepoRef{
		Owner: commit.Owner,
		Repo:  commit.Repo,
		Ref:   commit.Hash,
	}, checkRun, status); err != nil {
		return err
	}

	return nil
}

func (bw *BuildWorker) BuildProto(ctx context.Context, req *builder_tpb.BuildProtoMessage) (*emptypb.Empty, error) {

	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status: github.CheckRunStatusInProgress,
	}); err != nil {
		return nil, fmt.Errorf("check run: in progress: %w", err)
	}

	if req.Config.Git != nil {
		git.ExpandGitAliases(req.Config.Git, req.Commit)
	}

	source, err := bw.tmpClone(ctx, req.Commit)
	if err != nil {
		return nil, err
	}

	defer source.Close()

	if len(req.Config.ProtoBuilds) != 1 {
		return nil, fmt.Errorf("expected exactly one proto build")
	}
	buildSpec := req.Config.ProtoBuilds[0]

	logBuffer := &bytes.Buffer{}
	err = bw.buildProto(ctx, source, buildSpec.Name, logBuffer)

	if err != nil {
		if req.CheckRun == nil {
			return nil, fmt.Errorf("build: %w", err)
		}

		errorMessage := err.Error()
		fullText := fmt.Sprintf("%s\n\n```%s```", errorMessage, logBuffer.String())
		if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
			Status:     github.CheckRunStatusCompleted,
			Conclusion: some(github.CheckRunConclusionFailure),
			Output: &github.CheckRunOutput{
				Title:   some("proto build error"),
				Summary: errorMessage,
				Text:    some(fullText),
			},
		}); err != nil {
			log.Error(ctx, errorMessage)
			return nil, fmt.Errorf("build: update checkrun: failure: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	logStr := logBuffer.String()
	if len(logStr) >= 65535 {
		trunc := "... (truncated see logs for full error)"
		logStr = logStr[:65535-len(trunc)] + trunc
	}
	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status:     github.CheckRunStatusCompleted,
		Conclusion: some(github.CheckRunConclusionSuccess),
		Output: &github.CheckRunOutput{
			Title:   some("proto build success"),
			Summary: "proto build success",
			Text:    some(logStr),
		},
	}); err != nil {
		log.Error(ctx, logStr)
		return nil, fmt.Errorf("update checkrun: completed: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (bw *BuildWorker) buildProto(ctx context.Context, source builder.Source, builderName string, logWriter io.Writer) error {
	dest, err := newTmpDest()
	if err != nil {
		return fmt.Errorf("make tmp dest: %w", err)
	}

	buildSpec, err := source.PackageBuildConfig(builderName)
	if err != nil {
		return err
	}

	commitInfo, err := source.CommitInfo(ctx)
	if err != nil {
		return err
	}

	// Build
	if err := bw.builder.BuildProto(ctx, source, dest, builderName, io.MultiWriter(os.Stderr, logWriter)); err != nil {
		return err
	}

	// Package And Upload
	switch buildSpec.PackageType.(type) {
	case *config_j5pb.ProtoBuildConfig_GoProxy_:
		err := bw.uploadGoModule(ctx, commitInfo, dest.root)
		if err != nil {
			return fmt.Errorf("upload go module: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package type")
	}

	return nil
}

func (bw *BuildWorker) uploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, packageRoot string) error {

	gomodBytes, err := os.ReadFile(filepath.Join(packageRoot, "go.mod"))
	if err != nil {
		return err
	}

	parsedGoMod, err := modfile.Parse("go.mod", gomodBytes, nil)
	if err != nil {
		return err
	}

	if parsedGoMod.Module == nil {
		return fmt.Errorf("no module found in go.mod")
	}

	packageName := parsedGoMod.Module.Mod.Path

	commitHashPrefix := commitInfo.Hash
	if len(commitHashPrefix) > 12 {
		commitHashPrefix = commitHashPrefix[:12]
	}

	canonicalVersion := module.PseudoVersion("", "", commitInfo.Time.AsTime(), commitHashPrefix)

	log.WithFields(ctx, map[string]interface{}{
		"package": packageName,
		"version": canonicalVersion,
	}).Info("uploading go module")

	zipBuf := &bytes.Buffer{}

	err = zip.CreateFromDir(zipBuf, module.Version{
		Path:    packageName,
		Version: canonicalVersion,
	}, packageRoot)
	if err != nil {
		return err
	}

	metadata := map[string]string{
		gomodproxy.S3MetadataCommitTime: commitInfo.Time.AsTime().Format(time.RFC3339),
		gomodproxy.S3MetadataCommitHash: commitInfo.Hash,
	}

	dest := bw.goModUploader
	if err := dest.Put(ctx,
		path.Join(packageName, fmt.Sprintf("%s.mod", canonicalVersion)),
		strings.NewReader(string(gomodBytes)),
		metadata,
	); err != nil {
		return err
	}

	if err := dest.Put(ctx,
		path.Join(packageName, fmt.Sprintf("%s.zip", canonicalVersion)),
		zipBuf,
		metadata,
	); err != nil {
		return err
	}

	aliasMetadata := map[string]string{}
	for k, v := range metadata {
		aliasMetadata[k] = v
	}
	aliasMetadata[gomodproxy.S3MetadataAlias] = canonicalVersion

	for _, alias := range commitInfo.Aliases {
		if err := dest.Put(ctx,
			path.Join(packageName, fmt.Sprintf("%s.zip", alias)),
			bytes.NewReader([]byte(canonicalVersion)),
			aliasMetadata,
		); err != nil {
			return err
		}
	}

	if err := dest.Put(ctx,
		path.Join(packageName, fmt.Sprintf("%s.zip", commitInfo.Hash)),
		bytes.NewReader([]byte(canonicalVersion)),
		aliasMetadata,
	); err != nil {
		return err
	}

	return nil
}

func (bw *BuildWorker) BuildAPI(ctx context.Context, req *builder_tpb.BuildAPIMessage) (*emptypb.Empty, error) {

	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status: github.CheckRunStatusInProgress,
	}); err != nil {
		return nil, err
	}
	if req.Config.Git != nil {
		git.ExpandGitAliases(req.Config.Git, req.Commit)
	}

	source, err := bw.tmpClone(ctx, req.Commit)
	if err != nil {
		return nil, err
	}

	defer source.Close()

	if req.Config.Git != nil {
		git.ExpandGitAliases(req.Config.Git, req.Commit)
	}

	log.WithField(ctx, "commit", req.Commit).Info("Build API")

	err = bw.buildAPI(ctx, source)

	if err != nil {
		if req.CheckRun == nil {
			return nil, err
		}

		errorMessage := err.Error()
		if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
			Status:     github.CheckRunStatusCompleted,
			Conclusion: some(github.CheckRunConclusionFailure),
			Output: &github.CheckRunOutput{
				Title:   some("j5 error"),
				Summary: errorMessage,
				Text:    some(errorMessage),
			},
		}); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}

	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status:     github.CheckRunStatusCompleted,
		Conclusion: some(github.CheckRunConclusionSuccess),
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil

}

func (bw *BuildWorker) buildAPI(ctx context.Context, source builder.Source) error {

	config := source.J5Config()

	commitInfo, err := source.CommitInfo(ctx)
	if err != nil {
		return fmt.Errorf("commit info: %w", err)
	}

	packageName := path.Join(config.Registry.Organization, config.Registry.Name)

	img, err := source.SourceImage(ctx)
	if err != nil {
		return fmt.Errorf("source image: %w", err)
	}

	jdefDoc, err := structure.BuildFromImage(img)
	if err != nil {
		return fmt.Errorf("build from image: %w", err)
	}

	swaggerDoc, err := swagger.BuildSwagger(jdefDoc)
	if err != nil {
		return fmt.Errorf("build swagger: %w", err)
	}

	log.WithFields(ctx, map[string]interface{}{
		"package": packageName,
		"version": commitInfo.Hash,
		"aliases": commitInfo.Aliases,
	}).Info("uploading jsonapi")

	image, err := proto.Marshal(img)
	if err != nil {
		return err
	}

	jDefJSON, err := json.Marshal(jdefDoc)
	if err != nil {
		return err
	}

	swaggerJSON, err := json.Marshal(swaggerDoc)
	if err != nil {
		return err
	}

	versionDests := make([]string, 0, len(commitInfo.Aliases)+1)
	versionDests = append(versionDests, commitInfo.Hash)
	versionDests = append(versionDests, commitInfo.Aliases...)
	for _, version := range versionDests {
		p := path.Join(packageName, version)
		log.WithField(ctx, "path", p).Info("uploading image")

		if err := bw.j5Uploader.Put(ctx, path.Join(p, "image.bin"), bytes.NewReader(image), map[string]string{
			"Content-Type": "application/octet-stream",
		}); err != nil {
			return err
		}
		if err := bw.j5Uploader.Put(ctx, path.Join(p, "jdef.json"), bytes.NewReader(jDefJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
		if err := bw.j5Uploader.Put(ctx, path.Join(p, "swagger.json"), bytes.NewReader(swaggerJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
	}

	return nil
}

func some[T any](s T) *T {
	return &s
}

func (bw *BuildWorker) clone(ctx context.Context, commit *source_j5pb.CommitInfo, into string) error {

	ref := github.RepoRef{
		Owner: commit.Owner,
		Repo:  commit.Repo,
		Ref:   commit.Hash,
	}
	return bw.github.GetContent(ctx, ref, into)
}
