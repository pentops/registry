package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gomodproxy"
	"google.golang.org/protobuf/proto"
)

type FS interface {
	Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error
}

type NOPUploader struct {
}

func NewNOPUploader() *NOPUploader {
	return &NOPUploader{}
}

func (uu *NOPUploader) UploadGoModule(ctx context.Context, version FullInfo, goModData []byte, zipFile io.ReadCloser) error {
	return nil
}

func (uu *NOPUploader) UploadJsonAPI(ctx context.Context, info FullInfo, data J5Upload) error {
	return nil
}

type FSUploader struct {
	fs          FS
	GomodPrefix string
	JsonPrefix  string
}

func NewFSUploader(fs FS) *FSUploader {
	return &FSUploader{
		fs:          fs,
		GomodPrefix: "gomod",
		JsonPrefix:  "japi",
	}
}

type FullInfo struct {
	Version string
	Package string
	Commit  *builder_j5pb.CommitInfo
}

func (uu *FSUploader) UploadGoModule(ctx context.Context, version FullInfo, goModData []byte, zipFile io.ReadCloser) error {
	defer zipFile.Close()

	log.WithFields(ctx, map[string]interface{}{
		"package": version.Package,
		"version": version.Version,
	}).Info("uploading go module")

	metadata := map[string]string{
		gomodproxy.S3MetadataCommitTime: version.Commit.Time.AsTime().Format(time.RFC3339),
		gomodproxy.S3MetadataCommitHash: version.Commit.Hash,
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.mod", version.Version)),
		strings.NewReader(string(goModData)),
		metadata,
	); err != nil {
		return err
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", version.Version)),
		zipFile,
		metadata,
	); err != nil {
		return err
	}

	aliasMetadata := map[string]string{}
	for k, v := range metadata {
		aliasMetadata[k] = v
	}
	aliasMetadata[gomodproxy.S3MetadataAlias] = version.Version
	for _, alias := range version.Commit.Aliases {
		if err := uu.fs.Put(ctx,
			path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", alias)),
			bytes.NewReader([]byte(version.Version)),
			aliasMetadata,
		); err != nil {
			return err
		}
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", version.Commit.Hash)),
		bytes.NewReader([]byte(version.Version)),
		aliasMetadata,
	); err != nil {
		return err
	}

	return nil
}

func (uu *FSUploader) UploadJsonAPI(ctx context.Context, info FullInfo, data J5Upload) error {

	log.WithFields(ctx, map[string]interface{}{
		"package": info.Package,
		"version": info.Version,
		"aliases": info.Commit.Aliases,
	}).Info("uploading jsonapi")

	image, err := proto.Marshal(data.Image)
	if err != nil {
		return err
	}

	jDefJSON, err := json.Marshal(data.JDef)
	if err != nil {
		return err
	}

	swaggerJSON, err := json.Marshal(data.Swagger)
	if err != nil {
		return err
	}

	versionDests := make([]string, 0, len(info.Commit.Aliases)+1)
	versionDests = append(versionDests, info.Commit.Hash)
	versionDests = append(versionDests, info.Commit.Aliases...)
	for _, version := range versionDests {
		p := path.Join(uu.JsonPrefix, info.Package, version)
		log.WithField(ctx, "path", p).Info("uploading image")

		if err := uu.fs.Put(ctx, path.Join(p, "image.bin"), bytes.NewReader(image), map[string]string{
			"Content-Type": "application/octet-stream",
		}); err != nil {
			return err
		}
		if err := uu.fs.Put(ctx, path.Join(p, "jdef.json"), bytes.NewReader(jDefJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
		if err := uu.fs.Put(ctx, path.Join(p, "swagger.json"), bytes.NewReader(swaggerJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
	}

	return nil
}
