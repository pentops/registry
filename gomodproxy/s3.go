package gomodproxy

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pentops/registry/anyfs"
)

type S3PackageSrc struct {
	fs anyfs.FS
}

func NewS3PackageSrc(ctx context.Context, fs anyfs.FS) (*S3PackageSrc, error) {
	return &S3PackageSrc{
		fs: fs,
	}, nil
}

const (
	S3MetadataAlias      = "x-gomod-alias"
	S3MetadataCommitHash = "x-gomod-commit-hash"
	S3MetadataCommitTime = "x-gomod-commit-time"
)

func (src S3PackageSrc) s3Key(packageName, version, ext string) string {
	return path.Join(packageName, fmt.Sprintf("%s.%s", version, ext))
}

func (src *S3PackageSrc) Info(ctx context.Context, packageName, version string) (*Info, error) {

	key := src.s3Key(packageName, version, "zip")

	head, err := src.fs.Head(ctx, key)
	if err != nil {
		if errors.Is(err, anyfs.NotFoundError) {
			return nil, VersionNotFoundError(version)
		}
		return nil, err
	}

	alias, ok := head.Metadata[S3MetadataAlias]
	if ok && alias != version {
		// TODO: if alias then points back to itself then we have a loop
		return src.Info(ctx, packageName, alias)
	}

	commitTimeString, ok := head.Metadata[S3MetadataCommitTime]
	if !ok {
		return nil, fmt.Errorf("missing commit time")
	}

	commitTime, err := time.Parse(time.RFC3339, commitTimeString)
	if err != nil {
		return nil, err
	}

	return &Info{
		Version: version,
		Time:    commitTime,
	}, nil

}

func (src *S3PackageSrc) Latest(ctx context.Context, packageName string) (*Info, error) {
	return nil, NotImplementedError
}

func (src *S3PackageSrc) List(ctx context.Context, packageName string) ([]string, error) {

	versions := make([]string, 0)

	listOutput, err := src.fs.List(ctx, packageName)
	if err != nil {
		return nil, err
	}

	for _, obj := range listOutput {
		_, name := path.Split(obj.Name)
		if strings.HasSuffix(name, ".zip") {
			versions = append(versions, strings.TrimSuffix(name, ".zip"))
		}
	}

	return versions, nil
}

func (src *S3PackageSrc) modFromZip(ctx context.Context, packageName, canonicalVersion string) ([]byte, error) {
	scratchDir, err := os.MkdirTemp("", "gomodproxy")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(scratchDir)

	zipFileReader, zipFileSize, err := src.getZip(ctx, packageName, canonicalVersion)
	if err != nil {
		return nil, err
	}

	defer zipFileReader.Close()

	zipPath := filepath.Join(scratchDir, "zip.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(zipFile, zipFileReader); err != nil {
		zipFile.Close()
		return nil, err
	}
	if err := zipFile.Close(); err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(zipFile, zipFileSize)
	if err != nil {
		return nil, err
	}

	goMod, err := zipReader.Open("go.mod")
	if err != nil {
		return nil, err
	}

	defer goMod.Close()

	modBytes, err := io.ReadAll(goMod)
	if err != nil {
		return nil, err
	}

	return modBytes, nil

}

func (src *S3PackageSrc) Mod(ctx context.Context, packageName, version string) ([]byte, error) {
	canonical, err := src.Info(ctx, packageName, version)
	if err != nil {
		return nil, err
	}

	key := src.s3Key(packageName, canonical.Version, "mod")
	body, _, err := src.fs.GetBytes(ctx, key)
	if err != nil {
		if !errors.Is(err, anyfs.NotFoundError) {
			return nil, err
		}
		// Fallback.
		modBytes, err := src.modFromZip(ctx, packageName, canonical.Version)
		if err != nil {
			return nil, err
		}

		if err := src.fs.Put(ctx, key, bytes.NewReader(modBytes), map[string]string{
			S3MetadataCommitTime: canonical.Time.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}

		return modBytes, nil
	}

	return body, nil
}

func (src *S3PackageSrc) Zip(ctx context.Context, packageName, version string) (io.ReadCloser, error) {
	canonical, err := src.Info(ctx, packageName, version)
	if err != nil {
		return nil, err
	}

	reader, _, err := src.getZip(ctx, packageName, canonical.Version)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (src *S3PackageSrc) getZip(ctx context.Context, packageName string, canonicalVersion string) (io.ReadCloser, int64, error) {

	key := src.s3Key(packageName, canonicalVersion, "zip")
	obj, info, err := src.fs.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	return obj, info.Size, nil
}
