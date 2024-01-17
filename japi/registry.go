package japi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/jsonapi/swagger"
	"github.com/pentops/registry/anyfs"
	"google.golang.org/protobuf/proto"
	"gopkg.daemonl.com/log"
)

type FS interface {
	GetBytes(ctx context.Context, path string) ([]byte, *anyfs.FileInfo, error)
}

type Handler struct {
	Source FS
}

func NewRegistry(ctx context.Context, fs FS) (*Handler, error) {

	return &Handler{
		Source: fs,
	}, nil

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, fmt.Sprintf("path had %d parts, expected 5", len(parts)), http.StatusNotFound)
		return
	}

	orgName := parts[1]
	imageName := parts[2]
	version := parts[3]
	format := parts[4]

	ctx := r.Context()

	key := path.Join(orgName, imageName, version, "image.bin")
	ctx = log.WithFields(ctx, map[string]interface{}{
		"key": key,
	})

	bodyBytes, _, err := h.Source.GetBytes(ctx, key)
	if err != nil {
		if errors.Is(err, anyfs.NotFoundError) {
			http.NotFound(w, r)
			log.WithError(ctx, err).Error("not found")

			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	img := &source_j5pb.SourceImage{}
	if err := proto.Unmarshal(bodyBytes, img); err != nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch format {
	case "swagger.json":
		swaggerContent, err := buildSwagger(ctx, img)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(swaggerContent) // nolint: errcheck

	case "jdef.json":
		jdefContent, err := buildJDef(ctx, img)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jdefContent) // nolint: errcheck

	case "image.bin":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		imgBytes, err := proto.Marshal(img)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(imgBytes) // nolint: errcheck

	default:
		http.Error(w, fmt.Sprintf("unknown API format %s", format), http.StatusNotFound)
	}

}

func buildSwagger(ctx context.Context, img *source_j5pb.SourceImage) ([]byte, error) {
	jdefDoc, err := structure.BuildFromImage(img)
	if err != nil {
		return nil, err
	}

	swaggerDoc, err := swagger.BuildSwagger(jdefDoc)
	if err != nil {
		return nil, err
	}

	asJson, err := json.Marshal(swaggerDoc)
	if err != nil {
		return nil, err
	}

	return asJson, nil
}

func buildJDef(ctx context.Context, img *source_j5pb.SourceImage) ([]byte, error) {
	jdefDoc, err := structure.BuildFromImage(img)
	if err != nil {
		return nil, err
	}

	asJson, err := json.Marshal(jdefDoc)
	if err != nil {
		return nil, err
	}

	return asJson, nil
}
