package buildwrap

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/source"
	"github.com/pentops/registry/internal/github"
)

type tmpSource struct {
	//*source.Source
	commit *source_j5pb.CommitInfo
	dir    string
}

func (ts *tmpSource) close() error {
	return os.RemoveAll(ts.dir)
}

func (bw *BuildWorker) tmpClone(ctx context.Context, commit *source_j5pb.CommitInfo) (*tmpSource, error) {
	workDir, err := os.MkdirTemp("", "src")
	if err != nil {
		return nil, fmt.Errorf("make workdir: %w", err)
	}

	// Clone
	ref := github.RepoRef{
		Owner: commit.Owner,
		Repo:  commit.Repo,
		Ref:   commit.Hash,
	}
	if err := bw.github.GetContent(ctx, ref, workDir); err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("clone: %w", err)
	}

	return &tmpSource{
		dir:    workDir,
		commit: commit,
		//Source: buildSource,
	}, nil
}

func (src tmpSource) sourceForBundle(ctx context.Context, bundleRoot string) (*source.Source, error) {
	buildSource, err := source.ReadLocalSource(ctx, src.commit, os.DirFS(src.dir), bundleRoot)
	if err != nil {
		return nil, fmt.Errorf("build source: %w", err)
	}
	return buildSource, nil
}

type tmpDest struct {
	root string
	fs.FS
}

func newTmpDest() (*tmpDest, error) {
	dir, err := os.MkdirTemp("", "dst")
	if err != nil {
		return nil, fmt.Errorf("make tmp dir: %w", err)
	}

	return &tmpDest{
		root: dir,
		FS:   os.DirFS(dir),
	}, nil
}

func (local *tmpDest) Put(ctx context.Context, subPath string, body io.Reader) error {
	key := filepath.Join(local.root, subPath)
	err := os.MkdirAll(filepath.Dir(key), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(key)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err
	}

	return nil
}

func (td *tmpDest) Close() error {
	return os.RemoveAll(td.root)
}
