package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/bufbuild/protoyaml-go"
	"github.com/google/uuid"
	"github.com/pentops/flowtest"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/o5-deploy-aws/gen/o5/aws/deployer/v1/awsdeployer_tpb"
	"github.com/pentops/registry/gen/o5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/internal/gen/o5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_pb"
	"github.com/pentops/registry/internal/gen/o5/registry/github/v1/github_spb"
	"github.com/pentops/registry/internal/integration/mocks"
	"google.golang.org/protobuf/proto"
)

func TestO5Trigger(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	request := &awsdeployer_tpb.RequestDeploymentMessage{}
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

func mustMarshal(t flowtest.TB, pb proto.Message) string {
	t.Helper()
	b, err := protoyaml.Marshal(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return string(b)
}

func TestJ5Trigger(t *testing.T) {
	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	flow.Step("ConfigureRepo", func(ctx context.Context, t flowtest.Asserter) {
		res, err := uu.GithubCommand.ConfigureRepo(ctx, &github_spb.ConfigureRepoRequest{
			Owner: "owner",
			Name:  "repo",
			Config: &github_pb.RepoEventType_Configure{
				ChecksEnabled: true,
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

	var buildAPI *builder_tpb.BuildAPIMessage
	flow.Step("J5 Build", func(ctx context.Context, t flowtest.Asserter) {
		buildAPI = &builder_tpb.BuildAPIMessage{}

		uu.Github.TestPush("owner", "repo", mocks.GithubCommit{
			SHA: "after",
			Files: map[string]string{
				"j5.yaml": mustMarshal(t, &config_j5pb.Config{
					Bundles: []*config_j5pb.BundleConfig{{
						Dir: "bundle1",
					}, {
						Dir: "bundle2",
					}},
					Registry: &config_j5pb.RegistryConfig{
						Organization: "owner",
						Name:         "repo",
					},
				}),
				"bundle1/j5.yaml": mustMarshal(t, &config_j5pb.Config{
					Registry: nil,
				}),
				"bundle2/j5.yaml": mustMarshal(t, &config_j5pb.Config{
					Registry: &config_j5pb.RegistryConfig{
						Organization: "owner",
						Name:         "repo1",
					},
				}),
			},
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

		uu.Outbox.PopMessage(t, buildAPI)

		t.NotEmpty(buildAPI.Request)
		if buildAPI.Request == nil {
			t.Fatalf("unexpected nil request")
		}

		t.Equal([]string{".", "bundle2"}, buildAPI.Bundles)
	})

	flow.Step("J5 Reply", func(ctx context.Context, t flowtest.Asserter) {
		t.Logf("buildAPI: %v", buildAPI.Request)
		_, err := uu.BuilderReply.BuildStatus(ctx, &builder_tpb.BuildStatusMessage{
			Request: buildAPI.Request,
			Status:  builder_tpb.BuildStatus_BUILD_STATUS_SUCCESS,
		})
		t.NoError(err)

		gotStatus := uu.Github.CheckRunUpdates
		if len(gotStatus) != 1 {
			t.Fatalf("unexpected number of check runs: %d", len(gotStatus))
		}
		got := gotStatus[0]

		t.Equal("owner", got.CheckRun.Owner)
		t.Equal("repo", got.CheckRun.Repo)

	})

}
