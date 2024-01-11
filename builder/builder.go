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

	"github.com/google/uuid"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/log.go/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type IUploader interface {
	UploadGoModule(ctx context.Context, version FullInfo, goModData []byte, zipFile io.ReadCloser) error
	UploadJsonAPI(ctx context.Context, version FullInfo, jsonapiData []byte) error
}

type IDockerWrapper interface {
	Run(ctx context.Context, spec *jsonapi_pb.DockerSpec, input io.Reader, output io.Writer) error
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

func (b *Builder) BuildAll(ctx context.Context, spec *jsonapi_pb.Config, srcDir string, commitInfo *CommitInfo) error {
	protoBuildRequest, err := CodeGeneratorRequestFromSource(ctx, srcDir)
	if err != nil {
		return err
	}

	if spec.Git != nil {
		expandGitAliases(spec.Git, commitInfo)
	}

	for _, dockerBuild := range spec.ProtoBuilds {
		if err := b.BuildProto(ctx, srcDir, dockerBuild, protoBuildRequest, commitInfo); err != nil {
			return err
		}
	}

	log.Info(ctx, "build json API")

	image, err := structure.ReadImageFromSourceDir(ctx, srcDir)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	if err := b.Uploader.UploadJsonAPI(ctx, FullInfo{
		Package: path.Join(spec.Registry.Organization, spec.Registry.Name),
		Commit:  commitInfo,
	}, bb); err != nil {
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

func (b *Builder) BuildProto(ctx context.Context, srcDir string, dockerBuild *jsonapi_pb.ProtoBuildConfig, protoBuildRequest *pluginpb.CodeGeneratorRequest, commitInfo *CommitInfo) error {

	dest, err := os.MkdirTemp("", uuid.NewString())
	if err != nil {
		return err
	}
	packageRoot := filepath.Join(dest, "package")

	defer os.RemoveAll(dest)

	for _, plugin := range dockerBuild.Plugins {
		if err := b.RunProtocPlugin(ctx, packageRoot, plugin, protoBuildRequest); err != nil {
			return err
		}
	}

	switch pkg := dockerBuild.PackageType.(type) {
	case *jsonapi_pb.ProtoBuildConfig_GoProxy_:
		if err := copyFile(filepath.Join(srcDir, pkg.GoProxy.GoModFile), filepath.Join(packageRoot, "go.mod")); err != nil {
			return err
		}
		if err := b.PushGoPackage(ctx, dest, commitInfo); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported package type: %T", pkg)
	}

	return nil

}

func (b *Builder) RunProtocPlugin(ctx context.Context, dest string, plugin *jsonapi_pb.ProtoBuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest) error {

	if plugin.Label == "" {
		// This is a pretty poor way to label it, prefer spetting label
		// explicitly in config.
		plugin.Label = strings.Join([]string{
			plugin.Docker.Image, strings.Join(plugin.Docker.Entrypoint, ","), strings.Join(plugin.Docker.Command, ","),
		}, "/")
	}

	ctx = log.WithField(ctx, "builder", plugin.Label)
	log.Info(ctx, "running build plugin")

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

	log.Info(ctx, "build complete")

	return nil
}

func (b *Builder) PushGoPackage(ctx context.Context, root string, commitInfo *CommitInfo) error {
	packageRoot := filepath.Join(root, "package")

	gomodBytes, err := os.ReadFile(filepath.Join(packageRoot, "go.mod"))
	if err != nil {
		return err
	}

	parsedGoMod, err := modfile.Parse("go.mod", gomodBytes, nil)
	if err != nil {
		return err
	}

	if parsedGoMod.Module == nil {
		return fmt.Errorf("no module found in go.mod")
	}

	packageName := parsedGoMod.Module.Mod.Path

	commitHashPrefix := commitInfo.Hash
	if len(commitHashPrefix) > 12 {
		commitHashPrefix = commitHashPrefix[:12]
	}

	canonicalVersion := module.PseudoVersion("", "", commitInfo.Time, commitHashPrefix)

	zipFilePath := filepath.Join(root, canonicalVersion+".zip")

	if err := func() error {
		outWriter, err := os.Create(zipFilePath)
		if err != nil {
			return err
		}

		defer outWriter.Close()

		err = zip.CreateFromDir(outWriter, module.Version{
			Path:    packageName,
			Version: canonicalVersion,
		}, packageRoot)
		if err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return fmt.Errorf("creating zip file: %w", err)
	}

	info := FullInfo{
		Version: canonicalVersion,
		Package: packageName,
		Commit:  commitInfo,
	}

	zipReader, err := os.Open(zipFilePath)
	if err != nil {
		return err
	}

	return b.Uploader.UploadGoModule(ctx, info, gomodBytes, zipReader)

}
