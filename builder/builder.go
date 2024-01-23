package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/gen/j5/v1/schema_j5pb"
	"github.com/pentops/jsonapi/source"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/jsonapi/swagger"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type IUploader interface {
	BuildGoModule(ctx context.Context, commitInfo *builder_j5pb.CommitInfo, label string, callback BuilderCallback) error
	UploadJsonAPI(ctx context.Context, version FullInfo, jsonapiData J5Upload) error
}

type IDockerWrapper interface {
	Run(ctx context.Context, spec *source_j5pb.DockerSpec, input io.Reader, output io.Writer) error
}

type Builder struct {
	Docker   IDockerWrapper
	Uploader IUploader
}

func NewBuilder(docker IDockerWrapper, uploader IUploader) *Builder {
	return &Builder{
		Docker:   docker,
		Uploader: uploader,
	}
}

func (b *Builder) BuildAll(ctx context.Context, spec *source_j5pb.Config, srcDir string, commitInfo *builder_j5pb.CommitInfo, onlyMatching ...string) error {
	protoBuildRequest, err := CodeGeneratorRequestFromSource(ctx, srcDir)
	if err != nil {
		return err
	}

	if spec.Git != nil {
		expandGitAliases(spec.Git, commitInfo)
	}

	if len(onlyMatching) == 0 {
		if err := b.BuildJsonAPI(ctx, srcDir, spec.Registry, commitInfo); err != nil {
			return err
		}

		for _, dockerBuild := range spec.ProtoBuilds {
			if err := b.BuildProto(ctx, srcDir, dockerBuild, protoBuildRequest, commitInfo); err != nil {
				return err
			}
		}

		return nil
	}

	didAny := false

	for _, builderName := range onlyMatching {

		if builderName == "j5" {
			if err := b.BuildJsonAPI(ctx, srcDir, spec.Registry, commitInfo); err != nil {
				return err
			}
			didAny = true
			continue
		}

		subConfig := &source_j5pb.Config{
			Packages: spec.Packages,
			Options:  spec.Options,
			Registry: spec.Registry,
			Git:      spec.Git,
		}
		// format is proto/name, proto/name/plugin, j5
		if strings.HasPrefix(builderName, "proto/") {
			builderName = strings.TrimPrefix(builderName, "proto/")
			pluginName := ""
			if strings.Contains(builderName, "/") {
				parts := strings.SplitN(builderName, "/", 2)
				builderName = parts[0]
				pluginName = parts[1]
			}

			fmt.Printf("builderName: %s, pluginName: %s\n", builderName, pluginName)

			var foundProtoBuild *source_j5pb.ProtoBuildConfig
			for _, protoBuild := range spec.ProtoBuilds {
				if protoBuild.Label == builderName {
					foundProtoBuild = protoBuild
					break
				}
			}

			if foundProtoBuild == nil {
				return fmt.Errorf("proto build not found: %s", builderName)
			}

			if pluginName != "" {
				found := false
				for _, plugin := range foundProtoBuild.Plugins {
					if plugin.Label == pluginName {
						found = true
						foundProtoBuild.Plugins = []*source_j5pb.ProtoBuildPlugin{plugin}
						break
					}
				}

				if !found {
					return fmt.Errorf("plugin %s not found in proto builder %s", pluginName, builderName)
				}
			}

			subConfig.ProtoBuilds = []*source_j5pb.ProtoBuildConfig{foundProtoBuild}

			didAny = true
			if err := b.BuildProto(ctx, srcDir, foundProtoBuild, protoBuildRequest, commitInfo); err != nil {
				return err
			}

		}
	}

	if !didAny {
		return fmt.Errorf("no builders matched")
	}

	return nil
}

type J5Upload struct {
	Image   *source_j5pb.SourceImage
	JDef    *schema_j5pb.API
	Swagger *swagger.Document
}

func (b *Builder) BuildJsonAPI(ctx context.Context, srcDir string, registry *source_j5pb.RegistryConfig, commitInfo *builder_j5pb.CommitInfo) error {

	log.Info(ctx, "build json API")

	img, err := source.ReadImageFromSourceDir(ctx, srcDir)
	if err != nil {
		return err
	}

	jdefDoc, err := structure.BuildFromImage(img)
	if err != nil {
		return err
	}

	swaggerDoc, err := swagger.BuildSwagger(jdefDoc)
	if err != nil {
		return err
	}

	if err := b.Uploader.UploadJsonAPI(ctx, FullInfo{
		Package: path.Join(registry.Organization, registry.Name),
		Commit:  commitInfo,
	},

		J5Upload{
			Image:   img,
			JDef:    jdefDoc,
			Swagger: swaggerDoc,
		}); err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func (b *Builder) BuildProto(ctx context.Context, srcDir string, dockerBuild *source_j5pb.ProtoBuildConfig, protoBuildRequest *pluginpb.CodeGeneratorRequest, commitInfo *builder_j5pb.CommitInfo) error {

	switch pkg := dockerBuild.PackageType.(type) {
	case *source_j5pb.ProtoBuildConfig_GoProxy_:
		return b.Uploader.BuildGoModule(ctx, commitInfo, dockerBuild.Label, func(ctx context.Context, packageRoot string) error {
			for _, plugin := range dockerBuild.Plugins {
				if err := b.RunProtocPlugin(ctx, packageRoot, plugin, protoBuildRequest); err != nil {
					return err
				}
			}

			if err := copyFile(filepath.Join(srcDir, pkg.GoProxy.GoModFile), filepath.Join(packageRoot, "go.mod")); err != nil {
				return err
			}

			return nil
		})
	default:
		return fmt.Errorf("unsupported package type: %T", pkg)
	}

}

func (b *Builder) RunProtocPlugin(ctx context.Context, dest string, plugin *source_j5pb.ProtoBuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest) error {

	start := time.Now()
	if plugin.Label == "" {
		// This is a pretty poor way to label it, prefer spetting label
		// explicitly in config.
		plugin.Label = strings.Join([]string{
			plugin.Docker.Image, strings.Join(plugin.Docker.Entrypoint, ","), strings.Join(plugin.Docker.Command, ","),
		}, "/")
	}

	ctx = log.WithField(ctx, "builder", plugin.Label)
	log.Debug(ctx, "Running Protoc Plugin")

	parameter := strings.Join(plugin.Parameters, ",")
	sourceProto.Parameter = &parameter

	reqBytes, err := proto.Marshal(sourceProto)
	if err != nil {
		return err
	}

	resp := pluginpb.CodeGeneratorResponse{}

	outBuffer := &bytes.Buffer{}
	inBuffer := bytes.NewReader(reqBytes)
	err = b.Docker.Run(ctx, plugin.Docker, inBuffer, outBuffer)
	if err != nil {
		return fmt.Errorf("running docker %s: %w", plugin.Label, err)
	}

	if err := proto.Unmarshal(outBuffer.Bytes(), &resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("plugin error: %s", *resp.Error)
	}

	for _, f := range resp.File {
		name := f.GetName()
		fullPath := filepath.Join(dest, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(f.GetContent()), 0644); err != nil {
			return err
		}
	}

	log.WithFields(ctx, map[string]interface{}{
		"files":           len(resp.File),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil
}
