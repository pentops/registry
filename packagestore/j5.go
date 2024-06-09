package packagestore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path"
	"strings"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gen/o5/registry/registry/v1/registry_pb"
	"github.com/pentops/sqrlx.go/sqrlx"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *PackageStore) GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error) {

	ctx = log.WithFields(ctx, map[string]interface{}{
		"org":     orgName,
		"image":   imageName,
		"version": version,
	})
	log.Debug(ctx, "Getting J5 Image")

	pkg := &registry_pb.J5Package{}
	err := s.selectDataRow(ctx,
		sq.Select("data").
			From("j5_version").
			Where(sq.Eq{
				"owner":   orgName,
				"repo":    imageName,
				"version": version,
			}), pkg)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data, err := s.fs.GetBytes(ctx, pkg.StorageKey)
	if err != nil {
		return nil, err
	}

	img := &source_j5pb.SourceImage{}
	if err := protojson.Unmarshal(data, img); err != nil {
		return nil, err
	}

	return img, nil
}

func (s *PackageStore) UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage) error {

	packageName := path.Join(img.Registry.Organization, img.Registry.Name)

	log.WithFields(ctx, map[string]interface{}{
		"package": packageName,
		"owner":   commitInfo.Owner,
		"repo":    commitInfo.Repo,
		"version": commitInfo.Hash,
		"aliases": commitInfo.Aliases,
		"j5Org":   img.Registry.Organization,
		"j5Name":  img.Registry.Name,
	}).Info("uploading jsonapi")

	root := s.fs.Join("repo", commitInfo.Owner, commitInfo.Repo, "commit", commitInfo.Hash, "j5", packageName)

	imageBytes, err := protojson.Marshal(img)
	if err != nil {
		return err
	}

	storageKey := s.fs.Join(root, "image.json")

	if err := s.fs.Put(ctx, storageKey, bytes.NewReader(imageBytes), map[string]string{
		"Content-Type": "application/json",
	}); err != nil {
		return err
	}

	versionDests := make([]string, 0, len(commitInfo.Aliases)+1)
	versionDests = append(versionDests, commitInfo.Hash)

	for _, alias := range commitInfo.Aliases {
		versionDests = append(versionDests, strings.TrimPrefix(alias, "refs/heads/"))
	}

	pkg := &registry_pb.J5Package{
		Owner:      img.Registry.Organization,
		Name:       img.Registry.Name,
		Version:    commitInfo.Hash,
		StorageKey: storageKey,
		Aliases:    versionDests,
	}

	pkgJSON, err := protojson.Marshal(pkg)
	if err != nil {
		return err
	}

	log.WithFields(ctx, map[string]interface{}{
		"versions": versionDests,
	}).Info("Storing J5 Version")

	if err := s.db.Transact(ctx, &sqrlx.TxOptions{
		Isolation: sql.LevelReadCommitted,
		Retryable: true,
	}, func(ctx context.Context, tx sqrlx.Transaction) error {
		for _, version := range versionDests {
			_, err := tx.Insert(ctx, sqrlx.Upsert("j5_version").
				Key("owner", img.Registry.Organization).
				Key("repo", img.Registry.Name).
				Key("version", version).
				Set("data", pkgJSON))
			if err != nil {
				return err
			}

		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
