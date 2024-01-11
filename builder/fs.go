package builder

import (
	"context"

	git "github.com/go-git/go-git/v5"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gomodproxy"
)

func ExtractGitMetadata(ctx context.Context, dir string) (*gomodproxy.CommitInfo, error) {

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

		// TODO: Make this configurable
		if headName == "refs/heads/main" {
			commitAliases = append(commitAliases, "latest")
		}
	}

	// TODO: Tags

	log.WithFields(ctx, map[string]interface{}{
		"commitHash":    commitHash,
		"commitTime":    commitTime,
		"commitAliases": commitAliases,
	}).Info("Resolved Git Commit Info")

	return &gomodproxy.CommitInfo{
		Hash:    commitHash,
		Time:    commitTime,
		Aliases: commitAliases,
	}, nil
}
