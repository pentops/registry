package buildwrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/builder/git"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/github"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Storage interface {
	UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, fs fs.FS) error
	UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage) error
}

type BuildWorker struct {
	builder_tpb.UnimplementedBuilderTopicServer

	builder J5Builder
	github  IGithub

	store Storage
}

type J5Builder interface {
	BuildProto(ctx context.Context, source builder.Source, dest builder.FS, builderName string, logWriter io.Writer) error
}

type IGithub interface {
	GetContent(ctx context.Context, ref github.RepoRef, intoDir string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*source_j5pb.CommitInfo, error)
	UpdateCheckRun(ctx context.Context, ref github.RepoRef, checkRun *builder_tpb.CheckRun, status github.CheckRunUpdate) error
}

func NewBuildWorker(builder J5Builder, github IGithub, store Storage) *BuildWorker {

	return &BuildWorker{
		builder: builder,
		github:  github,
		store:   store,
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

	if req.CheckRun != nil {
		if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
			Status: github.CheckRunStatusInProgress,
		}); err != nil {
			return nil, fmt.Errorf("check run: in progress: %w", err)
		}
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

	if req.CheckRun != nil {
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
	}

	return &emptypb.Empty{}, nil
}

func (bw *BuildWorker) buildProto(ctx context.Context, source builder.Source, builderName string, logWriter io.Writer) error {
	dest, err := newTmpDest()
	if err != nil {
		return fmt.Errorf("make tmp dest: %w", err)
	}
	defer dest.Close()

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
		err := bw.store.UploadGoModule(ctx, commitInfo, dest)
		if err != nil {
			return fmt.Errorf("upload go module: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package type")
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

	commitInfo, err := source.CommitInfo(ctx)
	if err != nil {
		return fmt.Errorf("commit info: %w", err)
	}

	img, err := source.SourceImage(ctx)
	if err != nil {
		return fmt.Errorf("source image: %w", err)
	}

	return bw.store.UploadJ5Image(ctx, commitInfo, img)
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
