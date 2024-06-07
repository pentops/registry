package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/source"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-go/application/v1/application_pb"
	"github.com/pentops/o5-go/deployer/v1/deployer_tpb"
	"github.com/pentops/outbox.pg.go/outbox"
	"github.com/pentops/registry/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/github"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IClient interface {
	PullConfig(ctx context.Context, ref github.RepoRef, into proto.Message, tryPaths []string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*source_j5pb.CommitInfo, error)
	CreateCheckRun(ctx context.Context, ref github.RepoRef, name string, status *github.CheckRunUpdate) (int64, error)
}

type RefMatcher interface {
	GetRepo(ctx context.Context, push *github_pb.PushMessage) (*github_pb.RepoState, error)
}

type WebhookWorker struct {
	github    IClient
	refs      RefMatcher
	publisher Publisher

	github_pb.UnimplementedWebhookTopicServer
}

type Publisher interface {
	Publish(ctx context.Context, msg ...outbox.OutboxMessage) error
}

func NewWebhookWorker(refs RefMatcher, githubClient IClient, publisher Publisher) (*WebhookWorker, error) {
	return &WebhookWorker{
		github:    githubClient,
		publisher: publisher,
		refs:      refs,
	}, nil
}

func (ww *WebhookWorker) Push(ctx context.Context, event *github_pb.PushMessage) (*emptypb.Empty, error) {

	ctx = log.WithFields(ctx, map[string]interface{}{
		"owner":  event.Owner,
		"repo":   event.Repo,
		"ref":    event.Ref,
		"commit": event.After,
	})
	log.Debug(ctx, "Push")

	if !strings.HasPrefix(event.Ref, "refs/heads/") {
		log.Info(ctx, "Not a branch push, nothing to do")
	}

	branchName := strings.TrimPrefix(event.Ref, "refs/heads/")

	//repo := fmt.Sprintf("%s/%s", event.Owner, event.Repo)

	repo, err := ww.refs.GetRepo(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("target envs: %w", err)
	}
	if repo == nil {
		log.Info(ctx, "No repo config, nothing to do")
		return &emptypb.Empty{}, nil
	}

	targets := make([]*github_pb.DeployTargetType, 0, len(repo.Data.Branches))
	for _, target := range repo.Data.Branches {
		if target.BranchName == branchName {
			targets = append(targets, target.DeployTargets...)
		}
	}

	if len(targets) < 1 {
		log.Info(ctx, "No deploy targets, nothing to do")
		return &emptypb.Empty{}, nil
	}

	ref := github.RepoRef{
		Owner: event.Owner,
		Repo:  event.Repo,
		Ref:   event.After,
	}

	o5Envs := []string{}
	j5Build := false

	buildMessages := []outbox.OutboxMessage{}

	for _, target := range targets {
		switch target := target.Type.(type) {
		case *github_pb.DeployTargetType_O5Build_:
			o5Envs = append(o5Envs, target.O5Build.Environment)

		case *github_pb.DeployTargetType_J5Build_:
			j5Build = true
		default:
			return nil, fmt.Errorf("unknown target type: %T", target)
		}
	}

	if j5Build {
		builds, err := ww.j5Build(ctx, ref)
		if err != nil {
			checkRunError := &CheckRunError{}
			if repo.Data.ChecksEnabled && errors.As(err, checkRunError) {
				log.WithError(ctx, err).Error("j5 config error, reporting check run")
				_, err := ww.github.CreateCheckRun(ctx, ref, "j5-config", &github.CheckRunUpdate{
					Status:     github.CheckRunStatusCompleted,
					Conclusion: some(github.CheckRunConclusionFailure),
					Output: &github.CheckRunOutput{
						Title:   some(checkRunError.Title),
						Summary: checkRunError.Summary,
					},
				})
				if err != nil {
					return nil, fmt.Errorf("create check run: %w", err)
				}

			} else {
				return nil, fmt.Errorf("j5 build: %w", err)
			}
		}
		if builds.BuildAPI != nil {
			if repo.Data.ChecksEnabled {
				checkRunName := "j5-image"
				checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName, nil)
				if err != nil {
					return nil, fmt.Errorf("j5 image check run: %w", err)
				}

				builds.BuildAPI.CheckRun = &builder_tpb.CheckRun{
					Id:   checkRunID,
					Name: checkRunName,
				}
			}
			buildMessages = append(buildMessages, builds.BuildAPI)
		}

		for _, protoBuild := range builds.ProtoBuilds {
			if repo.Data.ChecksEnabled {
				checkRunName := fmt.Sprintf("j5-proto-%s", protoBuild.Config.ProtoBuilds[0].Name)
				checkRunID, err := ww.github.CreateCheckRun(ctx, ref, checkRunName, nil)
				if err != nil {
					return nil, fmt.Errorf("j5 proto check run: %w", err)
				}

				protoBuild.CheckRun = &builder_tpb.CheckRun{
					Id:   checkRunID,
					Name: checkRunName,
				}
			}
			buildMessages = append(buildMessages, protoBuild)

		}
	}

	if len(o5Envs) > 0 {
		builds, err := ww.o5Build(ctx, ref, o5Envs)
		if err != nil {
			return nil, fmt.Errorf("o5 build: %w", err)
		}

		for _, build := range builds {
			buildMessages = append(buildMessages, build)
		}
	}

	if len(buildMessages) > 0 {
		err := ww.publisher.Publish(ctx, buildMessages...)
		if err != nil {
			return nil, fmt.Errorf("publish: %w", err)
		}
	}

	return &emptypb.Empty{}, nil
}

var o5ConfigPaths = []string{
	"ext/o5/app.yaml",
	"ext/o5/app.yml",
	"o5.yaml",
	"o5.yml",
}

func (ww *WebhookWorker) o5Build(ctx context.Context, ref github.RepoRef, targetEnvs []string) ([]*deployer_tpb.RequestDeploymentMessage, error) {
	cfg := &application_pb.Application{}
	err := ww.github.PullConfig(ctx, ref, cfg, o5ConfigPaths)
	if err != nil {
		return nil, &CheckRunError{
			Title:   "o5 config error",
			Summary: err.Error(),
		}
	}

	triggers := make([]*deployer_tpb.RequestDeploymentMessage, 0, len(targetEnvs))

	for _, envID := range targetEnvs {
		triggers = append(triggers, &deployer_tpb.RequestDeploymentMessage{
			DeploymentId:  uuid.NewString(),
			Application:   cfg,
			Version:       ref.Ref,
			EnvironmentId: envID,
		})
	}

	return triggers, nil
}

type CheckRunError struct {
	Title   string
	Summary string
}

func (e CheckRunError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Summary)
}

type j5Buildset struct {
	BuildAPI    *builder_tpb.BuildAPIMessage
	ProtoBuilds []*builder_tpb.BuildProtoMessage
}

func (ww *WebhookWorker) j5Build(ctx context.Context, ref github.RepoRef) (*j5Buildset, error) {

	commitInfo, err := ww.github.GetCommit(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("get commit: %w", err)
	}

	ref.Ref = commitInfo.Hash

	cfg := &config_j5pb.Config{}
	err = ww.github.PullConfig(ctx, ref, cfg, source.ConfigPaths)
	if err != nil {
		log.WithError(ctx, err).Error("Config Error")
		return nil, &CheckRunError{
			Title:   "j5 config error",
			Summary: err.Error(),
		}
	}

	output := &j5Buildset{}

	{
		subConfig := &config_j5pb.Config{
			Packages: cfg.Packages,
			Options:  cfg.Options,
			Registry: cfg.Registry,
			Git:      cfg.Git,
		}

		req := &builder_tpb.BuildAPIMessage{
			Commit: commitInfo,
			Config: subConfig,
		}

		output.BuildAPI = req
	}

	for _, dockerBuild := range cfg.ProtoBuilds {
		subConfig := &config_j5pb.Config{
			ProtoBuilds: []*config_j5pb.ProtoBuildConfig{dockerBuild},
			Packages:    cfg.Packages,
			Options:     cfg.Options,
			Registry:    cfg.Registry,
			Git:         cfg.Git,
		}

		req := &builder_tpb.BuildProtoMessage{
			Commit: commitInfo,
			Config: subConfig,
		}

		output.ProtoBuilds = append(output.ProtoBuilds, req)

	}

	return output, nil
}

func some[T any](s T) *T {
	return &s
}
