package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/github"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/pluginpb"
)

type BuildWorker struct {
	builder_j5pb.UnimplementedBuilderTopicServer

	builder IBuilder
	github  IGithub
}

type IGithub interface {
	GetContent(ctx context.Context, ref github.RepoRef, intoDir string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*builder_j5pb.CommitInfo, error)
	UpdateCheckRun(ctx context.Context, ref github.RepoRef, checkRun *builder_j5pb.CheckRun, status github.CheckRunUpdate) error
}

type IBuilder interface {
	BuildProto(
		ctx context.Context,
		srcDir string,
		spec *source_j5pb.ProtoBuildConfig,
		generateRequest *pluginpb.CodeGeneratorRequest,
		commitInfo *builder_j5pb.CommitInfo,
	) error

	BuildJsonAPI(
		ctx context.Context,
		srcDir string,
		registry *source_j5pb.RegistryConfig,
		commitInfo *builder_j5pb.CommitInfo,
	) error
}

func NewBuildWorker(builder IBuilder, github IGithub) *BuildWorker {
	return &BuildWorker{
		builder: builder,
		github:  github,
	}
}

func (bw *BuildWorker) updateCheckRun(ctx context.Context, commit *builder_j5pb.CommitInfo, checkRun *builder_j5pb.CheckRun, status github.CheckRunUpdate) error {
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

func (bw *BuildWorker) BuildProto(ctx context.Context, req *builder_j5pb.BuildProtoMessage) (*emptypb.Empty, error) {

	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status: github.CheckRunStatusInProgress,
	}); err != nil {
		return nil, err
	}

	if req.Config.Git != nil {
		expandGitAliases(req.Config.Git, req.Commit)
	}

	workDir, err := os.MkdirTemp("", "src")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(workDir)

	// Clone
	err = bw.clone(ctx, req.Commit, workDir)
	if err != nil {
		return nil, err
	}

	if len(req.Config.ProtoBuilds) != 1 {
		return nil, fmt.Errorf("expected exactly one proto build")
	}
	buildSpec := req.Config.ProtoBuilds[0]

	// Build Request
	protoBuildRequest, err := CodeGeneratorRequestFromSource(ctx, workDir)
	if err != nil {
		if req.CheckRun == nil {
			return nil, err
		}

		errorMessage := err.Error()
		if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
			Status:     github.CheckRunStatusCompleted,
			Conclusion: some(github.CheckRunConclusionFailure),
			Output: &github.CheckRunOutput{
				Title:   some("proto request error"),
				Summary: errorMessage,
				Text:    some(errorMessage),
			},
		}); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}

	// Build
	if err := bw.builder.BuildProto(ctx, workDir, buildSpec, protoBuildRequest, req.Commit); err != nil {
		if req.CheckRun == nil {
			return nil, err
		}

		errorMessage := err.Error()
		if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
			Status:     github.CheckRunStatusCompleted,
			Conclusion: some(github.CheckRunConclusionFailure),
			Output: &github.CheckRunOutput{
				Title:   some("proto build error"),
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

func (bw *BuildWorker) BuildAPI(ctx context.Context, req *builder_j5pb.BuildAPIMessage) (*emptypb.Empty, error) {
	if err := bw.updateCheckRun(ctx, req.Commit, req.CheckRun, github.CheckRunUpdate{
		Status: github.CheckRunStatusInProgress,
	}); err != nil {
		return nil, err
	}

	if req.Config.Git != nil {
		expandGitAliases(req.Config.Git, req.Commit)
	}

	log.WithField(ctx, "commit", req.Commit).Info("Build API")

	workDir, err := os.MkdirTemp("", "src")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(workDir)

	// Clone
	err = bw.clone(ctx, req.Commit, workDir)
	if err != nil {
		return nil, err
	}

	// Build
	if err := bw.builder.BuildJsonAPI(ctx, workDir, req.Config.Registry, req.Commit); err != nil {
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

func some[T any](s T) *T {
	return &s
}

func (bw *BuildWorker) clone(ctx context.Context, commit *builder_j5pb.CommitInfo, into string) error {

	ref := github.RepoRef{
		Owner: commit.Owner,
		Repo:  commit.Repo,
		Ref:   commit.Hash,
	}
	return bw.github.GetContent(ctx, ref, into)
}