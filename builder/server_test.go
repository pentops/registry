package builder

import (
	"testing"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
)

func TestExpandGitAliases(t *testing.T) {

	cfg := &config_j5pb.GitConfig{
		Main: "refs/heads/main",
	}

	commitInfo := &builder_j5pb.CommitInfo{
		Aliases: []string{
			"refs/heads/main",
		},
	}

	expandGitAliases(cfg, commitInfo)

	gotAliases := map[string]bool{}
	for _, alias := range commitInfo.Aliases {
		gotAliases[alias] = true
	}

	wantAliases := map[string]bool{
		"main":   true,
		"latest": true,
	}

	for wantAlias := range wantAliases {
		if !gotAliases[wantAlias] {
			t.Errorf("missing alias %q", wantAlias)
		}
		delete(gotAliases, wantAlias)
	}

	for gotAlias := range gotAliases {
		t.Errorf("unexpected alias %q", gotAlias)
	}

}
