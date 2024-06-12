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
	"github.com/pentops/o5-deploy-aws/gen/o5/application/v1/application_pb"
	"github.com/pentops/o5-deploy-aws/gen/o5/aws/deployer/v1/awsdeployer_tpb"
	"github.com/pentops/o5-messaging/gen/o5/messaging/v1/messaging_pb"
	"github.com/pentops/o5-messaging/o5msg"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/internal/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/internal/github"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IClient interface {
	PullConfig(ctx context.Context, ref github.RepoRef, into proto.Message, tryPaths []string) error
	GetCommit(ctx context.Context, ref github.RepoRef) (*source_j5pb.CommitInfo, error)
	CreateCheckRun(ctx context.Context, ref github.RepoRef, name string, status *github.CheckRunUpdate) (*github_pb.CheckRun, error)
	UpdateCheckRun(ctx context.Context, checkRun *github_pb.CheckRun, status github.CheckRunUpdate) error
}

type RefMatcher interface {
	GetRepo(ctx context.Context, push *github_tpb.PushMessage) (*github_pb.RepoState, error)
}

type WebhookWorker struct {
	github    IClient
	refs      RefMatcher
	publisher Publisher

	github_tpb.UnimplementedWebhookTopicServer
	builder_tpb.UnimplementedBuilderReplyTopicServer
	awsdeployer_tpb.UnimplementedDeploymentReplyTopicServer
}

type Publisher interface {
	Publish(ctx context.Context, msg o5msg.Message) error
}

func NewWebhookWorker(refs RefMatcher, githubClient IClient, publisher Publisher) (*WebhookWorker, error) {
	return &WebhookWorker{
		github:    githubClient,
		publisher: publisher,
		refs:      refs,
	}, nil
}

func (ww *WebhookWorker) BuildStatus(ctx context.Context, message *builder_tpb.BuildStatusMessage) (*emptypb.Empty, error) {

	checkContext := &github_pb.CheckRun{}
	err := protojson.Unmarshal(message.Request.Context, checkContext)
	if err != nil {
		return nil, fmt.Errorf("unmarshal check context: %w", err)
	}

	status := github.CheckRunUpdate{}

	switch message.Status {
	case builder_tpb.BuildStatus_IN_PROGRESS:
		status.Status = github.CheckRunStatusInProgress

	case builder_tpb.BuildStatus_FAILURE:
		status.Status = github.CheckRunStatusCompleted
		status.Conclusion = some(github.CheckRunConclusionFailure)

	case builder_tpb.BuildStatus_SUCCESS:
		status.Status = github.CheckRunStatusCompleted
		status.Conclusion = some(github.CheckRunConclusionSuccess)
	}

	if message.Outcome != nil {
		status.Output = &github.CheckRunOutput{
			Title:   message.Outcome.Title,
			Summary: message.Outcome.Summary,
			Text:    message.Outcome.Text,
		}
	}

	if err := ww.github.UpdateCheckRun(ctx, checkContext, status); err != nil {
		return nil, fmt.Errorf("update check run: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (ww *WebhookWorker) githubCheck(ctx context.Context, ref github.RepoRef, checkRunName string) (*messaging_pb.RequestMetadata, error) {

	checkRun, err := ww.github.CreateCheckRun(ctx, ref, checkRunName, nil)
	if err != nil {
		return nil, fmt.Errorf("create check run: %w", err)
	}

	contextData, err := protojson.Marshal(checkRun)
	if err != nil {
		return nil, fmt.Errorf("marshal check run: %w", err)
	}

	return &messaging_pb.RequestMetadata{
		ReplyTo: "", // not filtered
		Context: contextData,
	}, nil
}

func (ww *WebhookWorker) Push(ctx context.Context, event *github_tpb.PushMessage) (*emptypb.Empty, error) {

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
		if target.BranchName == branchName || target.BranchName == "*" {
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

	buildMessages := []o5msg.Message{}

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
						Title:   checkRunError.Title,
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
				requestMetadata, err := ww.githubCheck(ctx, ref, "j5-image")
				if err != nil {
					return nil, fmt.Errorf("j5 image check run: %w", err)
				}
				builds.BuildAPI.Request = requestMetadata
			}
			buildMessages = append(buildMessages, builds.BuildAPI)
		}

		for _, protoBuild := range builds.ProtoBuilds {
			if repo.Data.ChecksEnabled {
				checkRunName := fmt.Sprintf("j5-proto-%s", protoBuild.Config.ProtoBuilds[0].Name)
				request, err := ww.githubCheck(ctx, ref, checkRunName)
				if err != nil {
					return nil, fmt.Errorf("j5 proto check run: %w", err)
				}
				protoBuild.Request = request
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
			if repo.Data.ChecksEnabled {
				requestMetadata, err := ww.githubCheck(ctx, ref, fmt.Sprintf("o5-deploy-%s", build.EnvironmentId))
				if err != nil {
					return nil, fmt.Errorf("o5 deploy check run: %w", err)
				}
				build.Request = requestMetadata
			}
			buildMessages = append(buildMessages, build)
		}
	}

	for _, msg := range buildMessages {
		err := ww.publisher.Publish(ctx, msg)
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

func (ww *WebhookWorker) o5Build(ctx context.Context, ref github.RepoRef, targetEnvs []string) ([]*awsdeployer_tpb.RequestDeploymentMessage, error) {
	cfg := &application_pb.Application{}
	err := ww.github.PullConfig(ctx, ref, cfg, o5ConfigPaths)
	if err != nil {
		return nil, &CheckRunError{
			Title:   "o5 config error",
			Summary: err.Error(),
		}
	}

	triggers := make([]*awsdeployer_tpb.RequestDeploymentMessage, 0, len(targetEnvs))

	for _, envID := range targetEnvs {
		triggers = append(triggers, &awsdeployer_tpb.RequestDeploymentMessage{
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
