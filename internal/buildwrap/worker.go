package buildwrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/pentops/j5/builder"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-messaging/gen/o5/messaging/v1/messaging_pb"
	"github.com/pentops/o5-messaging/o5msg"
	"github.com/pentops/registry/internal/gen/j5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/internal/github"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Storage interface {
	UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, fs fs.FS) error
	UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage, reg *config_j5pb.RegistryConfig) error
}

type Publisher interface {
	Publish(ctx context.Context, msg o5msg.Message) error
}

type BuildWorker struct {
	builder_tpb.UnimplementedBuilderRequestTopicServer

	builder J5Builder
	github  IGithub

	store     Storage
	publisher Publisher
}

type J5Builder interface {
	Publish(ctx context.Context, pc builder.PluginContext, input builder.Input, cfg *config_j5pb.PublishConfig) error
}

type IGithub interface {
	GetContent(ctx context.Context, ref github.RepoRef, intoDir string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*source_j5pb.CommitInfo, error)
}

func NewBuildWorker(builder J5Builder, github IGithub, store Storage, publisher o5msg.Publisher) *BuildWorker {
	return &BuildWorker{
		builder:   builder,
		github:    github,
		store:     store,
		publisher: publisher,
	}
}

func (bw *BuildWorker) RegisterGRPC(s *grpc.Server) {
	builder_tpb.RegisterBuilderRequestTopicServer(s, bw)
}

func (bw *BuildWorker) replyStatus(ctx context.Context, reply *messaging_pb.RequestMetadata, status builder_tpb.BuildStatus, outcome *builder_tpb.BuildOutcome) error {
	if reply == nil {
		return nil
	}
	return bw.publisher.Publish(ctx, &builder_tpb.BuildStatusMessage{
		Request: reply,
		Status:  status,
		Outcome: outcome,
	})
}

func (bw *BuildWorker) Publish(ctx context.Context, req *builder_tpb.PublishMessage) (*emptypb.Empty, error) {

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_IN_PROGRESS, nil); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
	}

	clone, err := bw.tmpClone(ctx, req.Commit)
	if err != nil {
		return nil, err
	}

	defer clone.close()

	logBuffer := &bytes.Buffer{}
	err = bw.runPublish(ctx, clone.fs(), req, logBuffer)

	if err != nil {
		if req.Request == nil {
			return nil, fmt.Errorf("build: %w", err)
		}

		errorMessage := err.Error()
		fullText := fmt.Sprintf("%s\n\n```%s```", errorMessage, logBuffer.String())
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_FAILURE, &builder_tpb.BuildOutcome{
			Title:   "proto build error",
			Summary: errorMessage,
			Text:    some(fullText),
		}); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	logStr := logBuffer.String()
	if len(logStr) >= 65535 {
		trunc := "... (truncated see logs for full error)"
		logStr = logStr[:65535-len(trunc)] + trunc
	}

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_SUCCESS, &builder_tpb.BuildOutcome{
			Title: "proto build success",
			Text:  some(logStr),
		}); err != nil {
			log.Error(ctx, logStr)
			return nil, fmt.Errorf("update checkrun: completed: %w", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (bw *BuildWorker) runPublish(ctx context.Context, sourceDir fs.FS, req *builder_tpb.PublishMessage, logBuffer io.Writer) error {

	dest, err := newTmpDest()
	if err != nil {
		return fmt.Errorf("make tmp dest: %w", err)
	}
	defer dest.Close()

	pc := builder.PluginContext{
		Variables: map[string]string{}, // TODO: Commit / Source Info
		ErrOut:    logBuffer,
		Dest:      dest,
	}

	input, err := builder.NewFSInput(ctx, sourceDir, req.Bundle)
	if err != nil {
		return fmt.Errorf("new fs input: %w", err)
	}

	cfg, err := input.J5Config()
	if err != nil {
		return fmt.Errorf("j5 config: %w", err)
	}

	var publishConfig *config_j5pb.PublishConfig
	for _, publish := range cfg.Publish {
		if publish.Name == req.Name {
			publishConfig = publish
			break
		}
	}
	if publishConfig == nil {
		return fmt.Errorf("publish config %q not found", req.Name)
	}

	// Build
	if err := bw.builder.Publish(ctx, pc, input, publishConfig); err != nil {
		return err
	}

	// Package And Upload
	if publishConfig.OutputFormat == nil || publishConfig.OutputFormat.Type == nil {
		return fmt.Errorf("output format not set")
	}
	switch publishConfig.OutputFormat.Type.(type) {
	case *config_j5pb.OutputType_GoProxy_:

		err = bw.store.UploadGoModule(ctx, req.Commit, dest)
		if err != nil {
			return fmt.Errorf("upload go module: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package type")
	}

	return nil
}

func (bw *BuildWorker) BuildAPI(ctx context.Context, req *builder_tpb.BuildAPIMessage) (*emptypb.Empty, error) {

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_IN_PROGRESS, nil); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
	}

	sourceClone, err := bw.tmpClone(ctx, req.Commit)
	if err != nil {
		return nil, err
	}

	defer sourceClone.close()

	log.WithField(ctx, "commit", req.Commit).Info("Build API")

	err = bw.buildAPI(ctx, sourceClone.fs(), req)

	if err != nil {
		if req.Request == nil {
			return nil, err
		}
		errorMessage := err.Error()
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_FAILURE, &builder_tpb.BuildOutcome{
			Title:   "proto build error",
			Summary: errorMessage,
		}); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_SUCCESS, &builder_tpb.BuildOutcome{
			Title: "API Build Success",
		}); err != nil {
			return nil, fmt.Errorf("update checkrun: completed: %w", err)
		}
	}

	return &emptypb.Empty{}, nil

}

func (bw *BuildWorker) buildAPI(ctx context.Context, sourceDir fs.FS, req *builder_tpb.BuildAPIMessage) error {

	input, err := builder.NewFSInput(ctx, sourceDir, req.Bundle)
	if err != nil {
		return fmt.Errorf("new fs input: %w", err)
	}

	bundleConfig, err := input.J5Config()
	if err != nil {
		return fmt.Errorf("j5 config: %w", err)
	}

	img, err := input.SourceImage(ctx)
	if err != nil {
		return fmt.Errorf("source image: %w", err)
	}

	return bw.store.UploadJ5Image(ctx, req.Commit, img, bundleConfig.Registry)
}

func some[T any](s T) *T {
	return &s
}
