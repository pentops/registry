package japi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/jdef"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/j5/schema/swagger"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
)

type ImageProvider interface {
	GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error)
}

func Handler(store ImageProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

		img, err := store.GetJ5Image(ctx, orgName, imageName, version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if img == nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			log.Error(ctx, "No Image")
			return
		}

		switch format {
		case "swagger.json":
			swaggerContent, err := buildSwagger(img)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(swaggerContent) // nolint: errcheck

		case "jdef.json":
			jdefContent, err := buildJDef(img)
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
	})

}

func buildSwagger(img *source_j5pb.SourceImage) ([]byte, error) {
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

func buildJDef(img *source_j5pb.SourceImage) ([]byte, error) {
	image, err := structure.BuildFromImage(img)
	if err != nil {
		return nil, err
	}

	jDefJSON, err := jdef.FromProto(image)
	if err != nil {
		return nil, err
	}

	jDefJSONBytes, err := json.Marshal(jDefJSON)
	if err != nil {
		return nil, err
	}

	return jDefJSONBytes, nil
}
