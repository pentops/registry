package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bufbuild/protoyaml-go"
	"github.com/google/uuid"
	"github.com/pentops/flowtest"
	"github.com/pentops/j5build/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/o5-deploy-aws/gen/o5/aws/deployer/v1/awsdeployer_tpb"
	"github.com/pentops/o5-messaging/outbox/outboxtest"
	"github.com/pentops/realms/j5auth"
	"github.com/pentops/registry/gen/j5/registry/builder/v1/builder_tpb"
	"github.com/pentops/registry/gen/j5/registry/github/v1/github_pb"
	"github.com/pentops/registry/gen/j5/registry/github/v1/github_spb"
	"github.com/pentops/registry/gen/j5/registry/github/v1/github_tpb"
	"github.com/pentops/registry/internal/integration/mocks"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func withTestActor(ctx context.Context) context.Context {

	jwt := map[string]interface{}{
		"sub": "test/" + uuid.NewString(),
	}
	jwtJSON, err := json.Marshal(jwt)
	if err != nil {
		panic(err)
	}

	md := metadata.MD{j5auth.VerifiedJWTHeader: []string{
		string(jwtJSON),
	}}

	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func TestO5Trigger(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	request := &awsdeployer_tpb.RequestDeploymentMessage{}
	environmentID := uuid.NewString()

	flow.Step("ConfigureRepo", func(ctx context.Context, t flowtest.Asserter) {
		ctx = withTestActor(ctx)
		res, err := uu.RepoCommand.ConfigureRepo(ctx, &github_spb.ConfigureRepoRequest{
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
			Commit: &github_pb.Commit{
				Owner: "owner",
				Repo:  "repo",
				Sha:   "after",
			},
			DeliveryId: uuid.NewString(),
			Ref:        "refs/heads/ref1",
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
		ctx = withTestActor(ctx)
		res, err := uu.RepoCommand.ConfigureRepo(ctx, &github_spb.ConfigureRepoRequest{
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

	var buildRoot *builder_tpb.BuildAPIMessage
	flow.Step("J5 Build", func(ctx context.Context, t flowtest.Asserter) {

		uu.Github.TestPush("owner", "repo", mocks.GithubCommit{
			SHA: "after",
			Files: map[string]string{
				"j5.yaml": mustMarshal(t, &config_j5pb.RepoConfigFile{
					Bundles: []*config_j5pb.BundleReference{{
						Dir:  "proto/b1",
						Name: "bundle1",
					}, {
						Dir:  "proto/b2",
						Name: "bundle2",
					}},
					Registry: &config_j5pb.RegistryConfig{
						Owner: "owner",
						Name:  "repo",
					},
				}),
				"proto/b1/j5.yaml": mustMarshal(t, &config_j5pb.BundleConfigFile{
					Registry: nil,
				}),
				"proto/b2/j5.yaml": mustMarshal(t, &config_j5pb.BundleConfigFile{
					Registry: &config_j5pb.RegistryConfig{
						Owner: "owner",
						Name:  "repo1",
					},
				}),
			},
		})

		_, err := uu.WebhookTopic.Push(ctx, &github_tpb.PushMessage{
			Commit: &github_pb.Commit{
				Owner: "owner",
				Repo:  "repo",
				Sha:   "after",
			},
			Ref:        "refs/heads/ref1",
			DeliveryId: uuid.NewString(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		buildRoot = &builder_tpb.BuildAPIMessage{}
		uu.Outbox.PopMessage(t, buildRoot, outboxtest.MessageBodyMatches(func(b *builder_tpb.BuildAPIMessage) bool {
			return b.Bundle == ""
		}))
		build2 := &builder_tpb.BuildAPIMessage{}
		uu.Outbox.PopMessage(t, build2, outboxtest.MessageBodyMatches(func(b *builder_tpb.BuildAPIMessage) bool {
			t.Logf("bundle: %q", b.Bundle)
			return b.Bundle == "bundle2"
		}))

		t.NotEmpty(buildRoot.Request)
		t.NotEmpty(build2.Request)

		t.Equal("", buildRoot.Bundle)
		t.Equal("bundle2", build2.Bundle)

	})

	flow.Step("J5 Reply", func(ctx context.Context, t flowtest.Asserter) {
		t.Logf("buildAPI: %v", buildRoot.Request)
		_, err := uu.BuilderReply.BuildStatus(ctx, &builder_tpb.BuildStatusMessage{
			Request: buildRoot.Request,
			Status:  builder_tpb.BuildStatus_BUILD_STATUS_SUCCESS,
		})
		t.NoError(err)

		gotStatus := uu.Github.CheckRunUpdates
		if len(gotStatus) != 1 {
			t.Fatalf("unexpected number of check runs: %d", len(gotStatus))
		}
		got := gotStatus[0]

		t.Equal("owner", got.CheckRun.CheckSuite.Commit.Owner)
		t.Equal("repo", got.CheckRun.CheckSuite.Commit.Repo)

	})

}
