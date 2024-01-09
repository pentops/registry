package gomodproxy

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/log.go/log"
)

type S3API interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, options ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, options ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, input *s3.HeadObjectInput, options ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, options ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3PackageSrc struct {
	bucket string
	prefix string
	client S3API
}

func NewS3PackageSrc(ctx context.Context, client S3API, location string) (*S3PackageSrc, error) {
	bucketURL, err := url.Parse(location)
	if err != nil {
		return nil, err
	}

	if bucketURL.Scheme != "s3" {
		return nil, fmt.Errorf("bucket must be an s3:// url")
	}

	bucketName := bucketURL.Host
	if bucketName == "" {
		return nil, fmt.Errorf("bucket must be an s3:// url")
	}

	keyPrefix := strings.TrimPrefix(bucketURL.Path, "/")

	log.WithFields(ctx, map[string]interface{}{
		"bucket": bucketName,
		"prefix": keyPrefix,
	}).Info("initializing s3 package source")

	return &S3PackageSrc{
		bucket: bucketName,
		prefix: keyPrefix,
		client: client,
	}, nil
}

const (
	S3MetadataAlias      = "x-gomod-alias"
	S3MetadataCommitHash = "x-gomod-commit-hash"
	S3MetadataCommitTime = "x-gomod-commit-time"
)

func (src S3PackageSrc) s3Key(packageName, version, ext string) string {
	return path.Join(src.prefix, packageName, fmt.Sprintf("%s.%s", version, ext))
}

func (src *S3PackageSrc) Info(ctx context.Context, packageName, version string) (*Info, error) {

	key := src.s3Key(packageName, version, "zip")

	log.WithFields(ctx, map[string]interface{}{
		"s3Bucket": src.bucket,
		"s3Key":    key,
	}).Info("fetching object")

	head, err := src.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &src.bucket,
		Key:    &key,
	})
	if err != nil {
		log.WithFields(ctx, map[string]interface{}{
			"s3Bucket": src.bucket,
			"s3Key":    key,
			"error":    err.Error(),
		}).Error("fetching object")

		if strings.Contains(err.Error(), "NotFound") {
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

	keyPrefix := path.Join(src.prefix, packageName)

	versions := make([]string, 0)

	var continuationToken *string

	for {
		listOutput, err := src.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &src.bucket,
			Prefix:            &keyPrefix,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("s3 list %s %s: %w", src.bucket, keyPrefix, err)
		}

		for _, obj := range listOutput.Contents {
			_, name := path.Split(*obj.Key)
			if strings.HasSuffix(name, ".zip") {
				versions = append(versions, strings.TrimSuffix(name, ".zip"))
			}
		}

		if listOutput.NextContinuationToken == nil {
			break
		}

		continuationToken = listOutput.NextContinuationToken
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
	obj, err := src.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &src.bucket,
		Key:    &key,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "NoSuchKey") {
			return nil, err
		}
		// Fallback.
		modBytes, err := src.modFromZip(ctx, packageName, canonical.Version)
		if err != nil {
			return nil, err
		}

		if err := src.put(ctx, key, bytes.NewReader(modBytes), map[string]string{
			S3MetadataCommitTime: canonical.Time.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}

		return modBytes, nil
	}

	defer obj.Body.Close()
	body, err := io.ReadAll(obj.Body)
	if err != nil {

		return nil, err
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
	obj, err := src.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &src.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, 0, err
	}

	return obj.Body, *obj.ContentLength, nil
}

func (src *S3PackageSrc) put(ctx context.Context, subPath string, body io.Reader, metadata map[string]string) error {
	key := path.Join(src.prefix, subPath)
	_, err := src.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   &src.bucket,
		Key:      &key,
		Body:     body,
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to s3: 's3://%s/%s' : %w", src.bucket, key, err)
	}
	return nil
}

type FullInfo struct {
	Version            string
	Time               time.Time
	OriginalCommitHash string
	Package            string
}

func (src *S3PackageSrc) UploadGoModule(ctx context.Context, version FullInfo, goModData []byte, zipFile io.ReadCloser) error {
	defer zipFile.Close()

	log.WithFields(ctx, map[string]interface{}{
		"package": version.Package,
		"version": version.Version,
	}).Info("uploading go module")

	metadata := map[string]string{
		S3MetadataCommitTime: version.Time.Format(time.RFC3339),
		S3MetadataCommitHash: version.OriginalCommitHash,
	}

	if err := src.put(ctx,
		path.Join(version.Package, fmt.Sprintf("%s.mod", version.Version)),
		strings.NewReader(string(goModData)),
		metadata,
	); err != nil {
		return err
	}

	if err := src.put(ctx,
		path.Join(version.Package, fmt.Sprintf("%s.zip", version.Version)),
		zipFile,
		metadata,
	); err != nil {
		return err
	}

	return nil
}
