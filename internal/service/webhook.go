package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5build/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-deploy-aws/gen/o5/application/v1/application_pb"
	"github.com/pentops/o5-deploy-aws/gen/o5/aws/deployer/v1/awsdeployer_tpb"
	"github.com/pentops/o5-messaging/o5msg"
	"github.com/pentops/registry/gen/j5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/gen/j5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/j5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/internal/git"
	"github.com/pentops/registry/internal/github"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IClient interface {
	PullConfig(ctx context.Context, ref *github_pb.Commit, into proto.Message, tryPaths []string) error
	GetCommit(ctx context.Context, ref *github_pb.Commit) (*source_j5pb.CommitInfo, error)
	CreateCheckRun(ctx context.Context, ref *github_pb.Commit, name string, status *github.CheckRunUpdate) (*github_pb.CheckRun, error)
	UpdateCheckRun(ctx context.Context, checkRun *github_pb.CheckRun, status github.CheckRunUpdate) error
}

type RefMatcher interface {
	GetRepo(ctx context.Context, owner string, name string) (*github_pb.RepoState, error)
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

func (ww *WebhookWorker) RegisterGRPC(srv *grpc.Server) {
	github_tpb.RegisterWebhookTopicServer(srv, ww)
	builder_tpb.RegisterBuilderReplyTopicServer(srv, ww)
	awsdeployer_tpb.RegisterDeploymentReplyTopicServer(srv, ww)
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

	log.WithFields(ctx, map[string]interface{}{
		"gh-status":  message.Status,
		"gh-outcome": message.Output,
	}).Debug("BuildStatus")

	if message.Output != nil {
		status.Output = &github.CheckRunOutput{
			Title:   message.Output.Title,
			Summary: message.Output.Summary,
			Text:    message.Output.Text,
		}
	}

	if err := ww.github.UpdateCheckRun(ctx, checkContext, status); err != nil {
		return nil, fmt.Errorf("update check run: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (ww *WebhookWorker) Push(ctx context.Context, event *github_tpb.PushMessage) (*emptypb.Empty, error) {
	ctx = log.WithFields(ctx, map[string]interface{}{
		"githubDeliveryID": event.DeliveryId,
		"owner":            event.Commit.Owner,
		"repo":             event.Commit.Repo,
		"commitSha":        event.Commit.Sha,
		"ref":              event.Ref,
	})
	log.Debug(ctx, "Push")

	if !strings.HasPrefix(event.Ref, "refs/heads/") {
		log.Info(ctx, "Not a branch push, nothing to do")
	}

	branchName := strings.TrimPrefix(event.Ref, "refs/heads/")
	return ww.kickOffChecks(ctx, event.Commit, branchName)
}

func (ww *WebhookWorker) CheckRun(ctx context.Context, event *github_tpb.CheckRunMessage) (*emptypb.Empty, error) {
	ctx = log.WithFields(ctx, map[string]interface{}{
		"githubDeliveryID": event.DeliveryId,
		"action":           event.Action,
		"owner":            event.CheckRun.CheckSuite.Commit.Owner,
		"repo":             event.CheckRun.CheckSuite.Commit.Repo,
		"commitSha":        event.CheckRun.CheckSuite.Commit.Sha,
		"branch":           event.CheckRun.CheckSuite.Branch,
		"checkRunId":       event.CheckRun.CheckId,
		"checkRunName":     event.CheckRun.CheckName,
		"checkSuiteId":     event.CheckRun.CheckSuite.CheckSuiteId,
	})
	log.Debug(ctx, "CheckRun")

	return &emptypb.Empty{}, nil
}

func (ww *WebhookWorker) CheckSuite(ctx context.Context, event *github_tpb.CheckSuiteMessage) (*emptypb.Empty, error) {
	ctx = log.WithFields(ctx, map[string]interface{}{
		"githubDeliveryID": event.DeliveryId,
		"owner":            event.CheckSuite.Commit.Owner,
		"repo":             event.CheckSuite.Commit.Repo,
		"branch":           event.CheckSuite.Branch,
		"commit":           event.CheckSuite.Commit.Sha,
		"suiteId":          event.CheckSuite.CheckSuiteId,
	})
	log.Debug(ctx, "CheckSuite")

	switch event.Action {
	case "requested", "rerequested":
		return ww.kickOffChecks(ctx, event.CheckSuite.Commit, event.CheckSuite.Branch)
	}
	return &emptypb.Empty{}, nil
}

func (ww *WebhookWorker) kickOffChecks(ctx context.Context, commit *github_pb.Commit, branchName string) (*emptypb.Empty, error) {
	buildTargets, repo, err := ww.buildTasksForBranch(ctx, commit, branchName)
	if err != nil {
		if !repo.Data.ChecksEnabled {
			return nil, err
		}
		checkRunError := &CheckRunError{}
		if !errors.As(err, checkRunError) {
			return nil, err
		}
		_, err = ww.github.CreateCheckRun(ctx, commit, checkRunError.RunName, &github.CheckRunUpdate{
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
		return &emptypb.Empty{}, nil
	}

	if repo.Data.ChecksEnabled {
		if err := ww.addGithubChecks(ctx, commit, buildTargets); err != nil {
			return nil, err
		}
	}

	if err := ww.publishTasks(ctx, buildTargets); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (ww *WebhookWorker) publishTasks(ctx context.Context, tasks []*buildTask) error {
	for _, task := range tasks {
		err := ww.publisher.Publish(ctx, task.message)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}
	return nil
}

func (ww *WebhookWorker) addGithubChecks(ctx context.Context, commit *github_pb.Commit, tasks []*buildTask) error {
	for _, task := range tasks {
		checkRun, err := ww.github.CreateCheckRun(ctx, commit, task.name, nil)
		if err != nil {
			return fmt.Errorf("create check run: %w", err)
		}
		contextData, err := protojson.Marshal(checkRun)
		if err != nil {
			return fmt.Errorf("marshal check run: %w", err)
		}
		task.message.SetJ5RequestMetadata(&messaging_j5pb.RequestMetadata{
			Context: contextData,
			// reply to not set.
		})
	}
	return nil
}

func (ww *WebhookWorker) buildTasksForBranch(ctx context.Context, commit *github_pb.Commit, branchName string) ([]*buildTask, *github_pb.RepoState, error) {
	repo, err := ww.refs.GetRepo(ctx, commit.Owner, commit.Repo)
	if err != nil {
		return nil, nil, fmt.Errorf("get repo: %w", err)
	}
	if repo == nil {
		log.Info(ctx, "No repo config, nothing to do")
		return nil, nil, nil
	}

	targets := make([]*github_pb.DeployTargetType, 0, len(repo.Data.Branches))
	for _, target := range repo.Data.Branches {
		if target.BranchName == branchName || target.BranchName == "*" {
			targets = append(targets, target.DeployTargets...)
		}
	}

	if len(targets) < 1 {
		log.Info(ctx, "No deploy targets, nothing to do")
		return nil, nil, nil
	}

	t2, err := ww.buildTargets(ctx, commit, targets)
	if err != nil {
		return nil, nil, fmt.Errorf("build targets: %w", err)
	}
	return t2, repo, nil
}

type taskMessage interface {
	o5msg.Message
	SetJ5RequestMetadata(*messaging_j5pb.RequestMetadata)
}

type buildTask struct {
	name    string
	message taskMessage
}

func (ww *WebhookWorker) buildTarget(ctx context.Context, commit *github_pb.Commit, target *github_pb.DeployTargetType) error {

	buildMessages, err := ww.buildTargets(ctx, commit, []*github_pb.DeployTargetType{target})
	if err != nil {
		return err
	}

	for _, msg := range buildMessages {
		err := ww.publisher.Publish(ctx, msg.message)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}

func (ww *WebhookWorker) buildTargets(ctx context.Context, commit *github_pb.Commit, targets []*github_pb.DeployTargetType) ([]*buildTask, error) {

	o5Envs := []string{}
	j5Build := false

	buildMessages := []*buildTask{}

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
		builds, err := ww.j5Build(ctx, commit)
		if err != nil {
			return nil, err

		}
		for _, apiBuild := range builds.APIBuilds {
			buildMessages = append(buildMessages, &buildTask{
				name:    "j5-image",
				message: apiBuild,
			})
		}

		for _, protoBuild := range builds.ProtoBuilds {
			checkRunName := fmt.Sprintf("j5-proto-%s", protoBuild.Name)
			buildMessages = append(buildMessages, &buildTask{
				name:    checkRunName,
				message: protoBuild,
			})
		}
	}

	if len(o5Envs) > 0 {
		builds, err := ww.o5Build(ctx, commit, o5Envs)
		if err != nil {
			return nil, fmt.Errorf("o5 build: %w", err)
		}

		for _, build := range builds {
			checkName := fmt.Sprintf("o5-deploy-%s", build.EnvironmentId)
			buildMessages = append(buildMessages, &buildTask{
				name:    checkName,
				message: build,
			})
		}
	}

	return buildMessages, nil
}

var o5ConfigPaths = []string{
	"ext/o5/app.yaml",
	"ext/o5/app.yml",
	"o5.yaml",
	"o5.yml",
}

func (ww *WebhookWorker) o5Build(ctx context.Context, commit *github_pb.Commit, targetEnvs []string) ([]*awsdeployer_tpb.RequestDeploymentMessage, *CheckRunError) {
	cfg := &application_pb.Application{}
	err := ww.github.PullConfig(ctx, commit, cfg, o5ConfigPaths)
	if err != nil {
		return nil, &CheckRunError{
			RunName: "o5-config",
			Title:   "o5 config error",
			Summary: err.Error(),
		}
	}

	triggers := make([]*awsdeployer_tpb.RequestDeploymentMessage, 0, len(targetEnvs))

	for _, envID := range targetEnvs {
		triggers = append(triggers, &awsdeployer_tpb.RequestDeploymentMessage{
			DeploymentId:  uuid.NewString(),
			Application:   cfg,
			Version:       commit.Sha,
			EnvironmentId: envID,
		})
	}

	return triggers, nil
}

type CheckRunError struct {
	RunName string
	Title   string
	Summary string
}

func (e CheckRunError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Summary)
}

type j5Buildset struct {
	APIBuilds   []*builder_tpb.BuildAPIMessage
	ProtoBuilds []*builder_tpb.PublishMessage
}

var configPaths = []string{
	"j5.yaml",
	"j5.repo.yaml",
	"ext/j5/j5.yaml",
}

var bundleConfigPaths = []string{
	"j5.yaml",
	"j5.bundle.yaml",
}

func (ww *WebhookWorker) j5Build(ctx context.Context, commit *github_pb.Commit) (*j5Buildset, error) {

	commitInfo, err := ww.github.GetCommit(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("get commit: %w", err)
	}

	commit.Sha = commitInfo.Hash

	cfg := &config_j5pb.RepoConfigFile{}
	err = ww.github.PullConfig(ctx, commit, cfg, configPaths)
	if err != nil {
		log.WithError(ctx, err).Error("Config Error")
		return nil, &CheckRunError{
			RunName: "j5-config",
			Title:   "j5 config error",
			Summary: err.Error(),
		}
	}

	if cfg.Git != nil {
		git.ExpandGitAliases(cfg.Git, commitInfo)
	}

	type namedBundle struct {
		name     string
		registry *config_j5pb.RegistryConfig
		publish  []*config_j5pb.PublishConfig
	}

	bundles := make([]namedBundle, 0, len(cfg.Bundles)+1)
	for _, bundle := range cfg.Bundles {
		bundleConfig := &config_j5pb.BundleConfigFile{}
		paths := make([]string, 0, len(bundleConfigPaths))
		for _, configPath := range bundleConfigPaths {
			paths = append(paths, path.Join(bundle.Dir, configPath))
		}
		if err := ww.github.PullConfig(ctx, commit, bundleConfig, paths); err != nil {
			return nil, &CheckRunError{
				RunName: "j5-config",
				Title:   "j5 bundle config error",
				Summary: fmt.Sprintf("Pulling %s/j5.yaml: %s", bundle.Dir, err.Error()),
			}
		}
		bundles = append(bundles, namedBundle{
			registry: bundleConfig.Registry,
			name:     bundle.Name,
			publish:  bundleConfig.Publish,
		})
	}

	if cfg.Registry != nil {
		// root is also a bundle.
		bundles = append(bundles, namedBundle{
			name:     "",
			registry: cfg.Registry,
			publish:  cfg.Publish,
		})

	}

	output := &j5Buildset{}

	for _, bundle := range bundles {
		if bundle.registry != nil {
			req := &builder_tpb.BuildAPIMessage{
				Commit: commitInfo,
				Bundle: bundle.name,
			}
			output.APIBuilds = append(output.APIBuilds, req)
		}

		for _, publish := range bundle.publish {
			req := &builder_tpb.PublishMessage{
				Commit: commitInfo,
				Name:   publish.Name,
				Bundle: bundle.name,
			}
			output.ProtoBuilds = append(output.ProtoBuilds, req)
		}

	}

	return output, nil
}

func some[T any](s T) *T {
	return &s
}
