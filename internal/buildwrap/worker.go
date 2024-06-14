package buildwrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-messaging/gen/o5/messaging/v1/messaging_pb"
	"github.com/pentops/o5-messaging/o5msg"
	"github.com/pentops/registry/internal/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/internal/github"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Storage interface {
	UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, fs fs.FS) error
	UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage) error
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
	BuildProto(ctx context.Context, source builder.Source, dest builder.FS, builderName string, logWriter io.Writer) error
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

func (bw *BuildWorker) BuildProto(ctx context.Context, req *builder_tpb.BuildProtoMessage) (*emptypb.Empty, error) {

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

	source, err := clone.sourceForBundle(ctx, ".") // TODO: Support sub-bundles for proto build.
	if err != nil {
		return nil, err
	}

	logBuffer := &bytes.Buffer{}
	err = bw.buildProto(ctx, source, req.Name, logBuffer)

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

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, builder_tpb.BuildStatus_IN_PROGRESS, nil); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
	}

	source, err := bw.tmpClone(ctx, req.Commit)
	if err != nil {
		return nil, err
	}

	defer source.close()

	log.WithField(ctx, "commit", req.Commit).Info("Build API")

	for _, bundle := range req.Bundles {
		subSource, err := source.sourceForBundle(ctx, bundle)
		if err != nil {
			return nil, err
		}

		err = bw.buildAPI(ctx, subSource)

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
