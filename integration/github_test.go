package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/bufbuild/protoyaml-go"
	"github.com/google/uuid"
	"github.com/pentops/flowtest"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/o5-go/deployer/v1/deployer_tpb"
	"github.com/pentops/registry/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/integration/mocks"
)

func TestO5Trigger(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	request := &deployer_tpb.RequestDeploymentMessage{}
	environmentID := uuid.NewString()

	flow.Step("ConfigureRepo", func(ctx context.Context, t flowtest.Asserter) {
		res, err := uu.GithubCommand.ConfigureRepo(ctx, &github_spb.ConfigureRepoRequest{
			Owner: "owner",
			Name:  "repo",
			Config: &github_pb.RepoEventType_Configure{
				ChecksEnabled: false,
				Branches: []*github_pb.Branch{{
					BranchName: "ref1",
					DeployTargets: []*github_pb.DeployTargetType{{
						Type: &github_pb.DeployTargetType_O5Build_{
							O5Build: &github_pb.DeployTargetType_O5Build{
								Environment: environmentID,
							},
						},
					}},
				}},
			},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		t.Equal("owner", res.Repo.Keys.Owner)

	})

	flow.Step("O5 Build", func(ctx context.Context, t flowtest.Asserter) {
		uu.Github.TestPush("owner", "repo", mocks.GithubCommit{
			SHA: "after",
			Files: map[string]string{
				"ext/o5/app.yaml": strings.Join([]string{
					"name: appname",
				}, "\n")},
		})

		_, err := uu.WebhookTopic.Push(ctx, &github_tpb.PushMessage{
			Owner: "owner",
			Repo:  "repo",
			Ref:   "refs/heads/ref1",
			After: "after",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		uu.Outbox.PopMessage(t, request)

		t.Equal(environmentID, request.EnvironmentId)
		t.Equal("appname", request.Application.Name)
		t.Equal("after", request.Version)

	})

}

func TestJ5Trigger(t *testing.T) {
	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	flow.Step("ConfigureRepo", func(ctx context.Context, t flowtest.Asserter) {
		res, err := uu.GithubCommand.ConfigureRepo(ctx, &github_spb.ConfigureRepoRequest{
			Owner: "owner",
			Name:  "repo",
			Config: &github_pb.RepoEventType_Configure{
				ChecksEnabled: false,
				Branches: []*github_pb.Branch{{
					BranchName: "ref1",
					DeployTargets: []*github_pb.DeployTargetType{{
						Type: &github_pb.DeployTargetType_J5Build_{
							J5Build: &github_pb.DeployTargetType_J5Build{},
						},
					}},
				}},
			},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		t.Equal("owner", res.Repo.Keys.Owner)

	})

	flow.Step("J5 Build", func(ctx context.Context, t flowtest.Asserter) {
		buildAPI := &builder_tpb.BuildAPIMessage{}

		cfg := &config_j5pb.Config{}

		cfgYaml, err := protoyaml.Marshal(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		uu.Github.TestPush("owner", "repo", mocks.GithubCommit{
			SHA:   "after",
			Files: map[string]string{"j5.yaml": string(cfgYaml)},
		})

		_, err = uu.WebhookTopic.Push(ctx, &github_tpb.PushMessage{
			Owner: "owner",
			Repo:  "repo",
			Ref:   "refs/heads/ref1",
			After: "after",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		uu.Outbox.PopMessage(t, buildAPI)

	})

}
