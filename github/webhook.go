package github

import (
	"context"
	"fmt"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/source"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-deploy-aws/gen/o5/github/v1/github_pb"
	"github.com/pentops/registry/messaging"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IClient interface {
	PullConfig(ctx context.Context, ref RepoRef, into proto.Message, tryPaths []string) error
	GetCommit(ctx context.Context, ref RepoRef) (*builder_j5pb.CommitInfo, error)
	CreateCheckRun(ctx context.Context, ref RepoRef, name string, status *CheckRunUpdate) (int64, error)
}

type Publisher interface {
	Publish(ctx context.Context, msg messaging.Message) error
}

type RefMatcher interface {
	PushTargets(*github_pb.PushMessage) []string
}

type WebhookWorker struct {
	github     IClient
	publisher  Publisher
	repos      []string
	checkRepos []string

	github_pb.UnimplementedWebhookTopicServer
}

func NewWebhookWorker(githubClient IClient, publisher Publisher, repos []string, checkRepos []string) (*WebhookWorker, error) {
	return &WebhookWorker{
		github:     githubClient,
		repos:      repos,
		publisher:  publisher,
		checkRepos: checkRepos,
	}, nil
}

func (ww *WebhookWorker) Push(ctx context.Context, event *github_pb.PushMessage) (*emptypb.Empty, error) {
	repo := fmt.Sprintf("%s/%s", event.Owner, event.Repo)

	matches := false
	skipChecks := true

	for _, r := range ww.repos {
		if r == repo {
			matches = true
			break
		}
	}

	for _, r := range ww.checkRepos {
		if r == repo {
			matches = true
			skipChecks = false
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
		return nil, fmt.Errorf("get commit: %w", err)
	}

	log.WithField(ctx, "commit", commitInfo).Debug("Got commit")

	ref.Ref = commitInfo.Hash

	cfg := &source_j5pb.Config{}
	err = ww.github.PullConfig(ctx, ref, cfg, source.ConfigPaths)
	if err != nil {
		log.WithError(ctx, err).Error("Config Error")
		_, err := ww.github.CreateCheckRun(ctx, ref, "j5-config", &CheckRunUpdate{
			Status:     CheckRunStatusCompleted,
			Conclusion: some(CheckRunConclusionFailure),
			Output: &CheckRunOutput{
				Title:   some("j5 config error"),
				Summary: err.Error(),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("j5 config check run: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	{
		subConfig := &source_j5pb.Config{
			Packages: cfg.Packages,
			Options:  cfg.Options,
			Registry: cfg.Registry,
			Git:      cfg.Git,
		}

		req := &builder_j5pb.BuildAPIMessage{
			Commit: commitInfo,
			Config: subConfig,
		}

		if !skipChecks {
			checkRunName := "j5-image"
			checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName, nil)
			if err != nil {
				return nil, fmt.Errorf("j5 image check run: %w", err)
			}

			req.CheckRun = &builder_j5pb.CheckRun{
				Id:   checkRunID,
				Name: checkRunName,
			}
		}

		err = ww.publisher.Publish(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("publish: j5 image: %w", err)
		}
	}

	for _, dockerBuild := range cfg.ProtoBuilds {
		log.Debug(ctx, "Publishing docker build")
		subConfig := &source_j5pb.Config{
			ProtoBuilds: []*source_j5pb.ProtoBuildConfig{dockerBuild},
			Packages:    cfg.Packages,
			Options:     cfg.Options,
			Registry:    cfg.Registry,
			Git:         cfg.Git,
		}

		req := &builder_j5pb.BuildProtoMessage{
			Commit: commitInfo,
			Config: subConfig,
		}

		if !skipChecks {
			checkRunName := fmt.Sprintf("j5-proto-%s", dockerBuild.Label)
			checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName, nil)
			if err != nil {
				return nil, fmt.Errorf("j5 proto check run: %w", err)
			}

			req.CheckRun = &builder_j5pb.CheckRun{
				Id:   checkRunID,
				Name: checkRunName,
			}
		}

		err = ww.publisher.Publish(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("publish: j5 proto: %w", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func some[T any](s T) *T {
	return &s
}
