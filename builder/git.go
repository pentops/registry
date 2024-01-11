package builder

import (
	"context"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"github.com/pentops/log.go/log"
)

func expandGitAliases(gitConfig *jsonapi_pb.GitConfig, commitInfo *CommitInfo) {
	for _, alias := range commitInfo.Aliases {
		if strings.HasPrefix(alias, "refs/tags/") {
			commitInfo.Aliases = append(commitInfo.Aliases, strings.TrimPrefix(alias, "refs/tags/"))
		} else if strings.HasPrefix(alias, "refs/heads/") {
			branchName := strings.TrimPrefix(alias, "refs/heads/")
			commitInfo.Aliases = append(commitInfo.Aliases, branchName)
		}
		if globMatch(alias, commitInfo.Hash) {
			commitInfo.Aliases = append(commitInfo.Aliases, alias)
		}
	}
}

func ExtractGitMetadata(ctx context.Context, gitConfig *jsonapi_pb.GitConfig, dir string) (*CommitInfo, error) {

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	commitTime := commit.Committer.When
	commitHash := commit.Hash.String()

	var commitAliases []string
	commitAliases = append(commitAliases, commitHash)

	headName := head.Name()
	if headName.IsBranch() {
		commitAliases = append(commitAliases, headName.Short())
		commitAliases = append(commitAliases, string(headName))

		if globMatch(gitConfig.Main, string(headName)) {
			commitAliases = append(commitAliases, "latest")
		}
	}

	// TODO: Tags, including latest match on /refs/tags/v* or whatever is
	// configured

	log.WithFields(ctx, map[string]interface{}{
		"commitHash":    commitHash,
		"commitTime":    commitTime,
		"commitAliases": commitAliases,
	}).Info("Resolved Git Commit Info")

	return &CommitInfo{
		Hash:    commitHash,
		Time:    commitTime,
		Aliases: commitAliases,
	}, nil
}
