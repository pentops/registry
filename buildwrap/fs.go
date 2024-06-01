package buildwrap

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/schema/source"
)

type RemoteWithMetadata interface {
	Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error
}

type PrefixedRemote struct {
	remote RemoteWithMetadata
	prefix string
}

func SubRemote(rr RemoteWithMetadata, prefix string) *PrefixedRemote {
	return &PrefixedRemote{
		remote: rr,
		prefix: prefix,
	}
}

func (pr *PrefixedRemote) Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error {
	return pr.remote.Put(ctx, filepath.Join(pr.prefix, path), body, metadata)
}

type tmpSource struct {
	*source.Source
	dir string
}

func (ts *tmpSource) Close() error {
	return os.RemoveAll(ts.dir)
}

func (bw *BuildWorker) tmpClone(ctx context.Context, commit *source_j5pb.CommitInfo) (*tmpSource, error) {
	workDir, err := os.MkdirTemp("", "src")
	if err != nil {
		return nil, fmt.Errorf("make workdir: %w", err)
	}

	// Clone
	err = bw.clone(ctx, commit, workDir)
	if err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("clone: %w", err)
	}

	buildSource, err := source.ReadLocalSource(ctx, commit, workDir)
	if err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("build source: %w", err)
	}

	return &tmpSource{
		dir:    workDir,
		Source: buildSource,
	}, nil
}

type tmpDest struct {
	root string
}

func newTmpDest() (*tmpDest, error) {
	dir, err := os.MkdirTemp("", "dst")
	if err != nil {
		return nil, fmt.Errorf("make tmp dir: %w", err)
	}

	return &tmpDest{
		root: dir,
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
