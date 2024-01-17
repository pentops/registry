package builder

import (
	"context"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/glob"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func expandGitAliases(gitConfig *source_j5pb.GitConfig, commitInfo *builder_j5pb.CommitInfo) {
	aliases := make([]string, 0, len(commitInfo.Aliases))
	for _, alias := range commitInfo.Aliases {
		if strings.HasPrefix(alias, "refs/tags/") {
			aliases = append(aliases, strings.TrimPrefix(alias, "refs/tags/"))
		} else if strings.HasPrefix(alias, "refs/heads/") {
			branchName := strings.TrimPrefix(alias, "refs/heads/")
			aliases = append(aliases, branchName)
		} else {
			aliases = append(aliases, alias)
		}
		if glob.GlobMatch(gitConfig.Main, alias) {
			aliases = append(aliases, "latest")
		}
	}
	commitInfo.Aliases = aliases
}

func ExtractGitMetadata(ctx context.Context, gitConfig *source_j5pb.GitConfig, dir string) (*builder_j5pb.CommitInfo, error) {

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

		if glob.GlobMatch(gitConfig.Main, string(headName)) {
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

	info := &builder_j5pb.CommitInfo{
		Hash:    commitHash,
		Time:    timestamppb.New(commitTime),
		Aliases: commitAliases,
	}

	/* TODO: pull this from the repo config
	origin, err := repo.Remote("origin")
	if err == nil {
		url := origin.Config().URLs[0]
		info.Owner = origin.Config().URLs[0]
		info.Repo = origin.Config().URLs[0]
	}*/

	return info, nil
}
