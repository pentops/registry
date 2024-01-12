package github

import (
	"context"
	"fmt"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-go/github/v1/github_pb"
	"github.com/pentops/registry/messaging"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IClient interface {
	PullConfig(ctx context.Context, ref RepoRef, into proto.Message, tryPaths []string) error
	GetCommit(ctx context.Context, ref RepoRef) (*builder_j5pb.CommitInfo, error)
	CreateCheckRun(ctx context.Context, ref RepoRef, name string) (int64, error)
}

type Publisher interface {
	Publish(ctx context.Context, msg messaging.Message) error
}

type RefMatcher interface {
	PushTargets(*github_pb.PushMessage) []string
}

type WebhookWorker struct {
	github    IClient
	publisher Publisher
	repos     []string

	github_pb.UnimplementedWebhookTopicServer
}

func NewWebhookWorker(githubClient IClient, publisher Publisher, repos []string) (*WebhookWorker, error) {
	return &WebhookWorker{
		github:    githubClient,
		repos:     repos,
		publisher: publisher,
	}, nil
}

func (ww *WebhookWorker) Push(ctx context.Context, event *github_pb.PushMessage) (*emptypb.Empty, error) {

	repo := fmt.Sprintf("%s/%s", event.Owner, event.Repo)

	matches := false
	for _, r := range ww.repos {
		if r == repo {
			matches = true
			break
		}
	}
	if !matches {
		log.Info(ctx, "Repo not configured, nothing to do")
		return &emptypb.Empty{}, nil
	}

	ctx = log.WithFields(ctx, map[string]interface{}{
		"owner":  event.Owner,
		"repo":   event.Repo,
		"ref":    event.Ref,
		"commit": event.After,
	})
	log.Debug(ctx, "Push")

	ref := RepoRef{
		Owner: event.Owner,
		Repo:  event.Repo,
		Ref:   event.After,
	}

	commitInfo, err := ww.github.GetCommit(ctx, ref)
	if err != nil {
		return nil, err
	}

	ref.Ref = commitInfo.Hash

	cfg := &config_j5pb.Config{}
	err = ww.github.PullConfig(ctx, ref, cfg, []string{
		"j5.yaml",
		"jsonapi.yaml",
		"j5.yml",
		"jsonapi.yml",
		"ext/j5/j5.yaml",
		"ext/j5/j5.yml",
	})
	if err != nil {
		return nil, err
	}

	{
		subConfig := &config_j5pb.Config{
			Packages: cfg.Packages,
			Options:  cfg.Options,
			Registry: cfg.Registry,
			Git:      cfg.Git,
		}

		checkRunName := "j5-image"
		checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName)
		if err != nil {
			return nil, err
		}

		req := &builder_j5pb.BuildAPIMessage{
			Commit: commitInfo,
			Config: subConfig,
			CheckRun: &builder_j5pb.CheckRun{
				Id:   checkRunID,
				Name: checkRunName,
			},
		}

		err = ww.publisher.Publish(ctx, req)
		if err != nil {
			return nil, err
		}

	}

	for _, dockerBuild := range cfg.ProtoBuilds {
		log.Debug(ctx, "Publishing docker build")
		subConfig := &config_j5pb.Config{
			ProtoBuilds: []*config_j5pb.ProtoBuildConfig{dockerBuild},
			Packages:    cfg.Packages,
			Options:     cfg.Options,
			Registry:    cfg.Registry,
			Git:         cfg.Git,
		}

		checkRunName := fmt.Sprintf("j5-proto-%s", dockerBuild.Label)
		checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName)
		if err != nil {
			return nil, err
		}
		req := &builder_j5pb.BuildProtoMessage{
			Commit: commitInfo,
			Config: subConfig,
			CheckRun: &builder_j5pb.CheckRun{
				Id:   checkRunID,
				Name: checkRunName,
			},
		}
		err = ww.publisher.Publish(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}
