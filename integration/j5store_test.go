package integration

import (
	"context"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestJ5Store(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	commitHash := "we5rbvfcb"
	flow.Step("Upload J5 Config", func(ctx context.Context, t flowtest.Asserter) {
		if err := uu.PackageStore.UploadJ5Image(ctx, &source_j5pb.CommitInfo{
			Owner: "commitorg",
			Repo:  "commitrepo",
			Hash:  commitHash,
			Time:  timestamppb.Now(),
			Aliases: []string{
				"refs/heads/main",
			},
		}, &source_j5pb.SourceImage{
			Registry: &config_j5pb.RegistryConfig{
				Organization: "cfgorg",
				Name:         "cfgrepo",
			},
			File: []*descriptorpb.FileDescriptorProto{{
				Name:    proto.String("test/v1/test.proto"),
				Package: proto.String("test.v1"),
			}},
			Codec: &config_j5pb.CodecOptions{},
			Packages: []*config_j5pb.PackageConfig{{
				Name:  "test.v1",
				Label: "Test Package",
			}},
		}); err != nil {
			t.Fatalf("failed to upload j5 image: %v", err)
		}
	})

	flow.Step("Download By Branch", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/registry/v1/cfgorg/cfgrepo/main/jdef.json")
		t.Equal(200, res.StatusCode)
		t.Log(string(res.Body))

		aa := jsontest.NewTestAsserter(t, res.Body)

		aa.AssertEqual("packages.0.name", "test.v1")
	})
	flow.Step("Download By Commit", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/registry/v1/cfgorg/cfgrepo/"+commitHash+"/jdef.json")
		t.Equal(200, res.StatusCode)
		t.Log(string(res.Body))

		aa := jsontest.NewTestAsserter(t, res.Body)

		aa.AssertEqual("packages.0.name", "test.v1")
	})
}
