package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/gomodproxy"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type BufGenerateConfig struct {
	Version string                      `json:"version,omitempty" yaml:"version,omitempty"`
	Plugins []BufGeneratePluginConfigV1 `json:"plugins,omitempty" yaml:"plugins,omitempty"`
}

// ExternalPluginConfigV1 is an external plugin configuration.
type BufGeneratePluginConfigV1 struct {
	Plugin     string      `json:"plugin,omitempty" yaml:"plugin,omitempty"`
	Revision   int         `json:"revision,omitempty" yaml:"revision,omitempty"`
	Name       string      `json:"name,omitempty" yaml:"name,omitempty"`
	Remote     string      `json:"remote,omitempty" yaml:"remote,omitempty"`
	Out        string      `json:"out,omitempty" yaml:"out,omitempty"`
	Opt        interface{} `json:"opt,omitempty" yaml:"opt,omitempty"`
	Path       string      `json:"path,omitempty" yaml:"path,omitempty"`
	ProtocPath string      `json:"protoc_path,omitempty" yaml:"protoc_path,omitempty"`
	Strategy   string      `json:"strategy,omitempty" yaml:"strategy,omitempty"`

	Docker *DockerConfig `json:"docker,omitempty" yaml:"docker,omitempty"`
}

type DockerConfig struct {
	Image      string      `json:"image,omitempty" yaml:"image,omitempty"`
	Env        []string    `json:"env,omitempty" yaml:"env,omitempty"`
	Entrypoint []string    `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Command    []string    `json:"command,omitempty" yaml:"command,omitempty"`
	Pull       *PullConfig `json:"pull,omitempty" yaml:"pull,omitempty"`
}

func ConvertBufGenerateSpec(src *BufGenerateConfig) ([]DockerSpec, error) {

	out := make([]DockerSpec, 0, len(src.Plugins))

	for _, plugin := range src.Plugins {

		if plugin.Docker == nil {
			return nil, fmt.Errorf("plugins require Docker spec")
		}

		if plugin.Strategy != "" {
			return nil, fmt.Errorf("unsupported strategy: %s", plugin.Strategy)
		}

		if plugin.ProtocPath != "" {
			return nil, fmt.Errorf("unsupported protoc_path: %s", plugin.ProtocPath)
		}

		env := make([]string, len(plugin.Docker.Env))
		for idx, src := range plugin.Docker.Env {
			parts := strings.SplitN(src, "=", 1)
			if len(parts) == 1 {
				env[idx] = fmt.Sprintf("%s=%s", src, os.Getenv(src))
				continue
			}
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid env var: %s", src)
			}
			val := os.ExpandEnv(src)
			env[idx] = fmt.Sprintf("%s=%s", parts[0], val)
		}

		spec := DockerSpec{
			Image:      plugin.Docker.Image,
			Name:       plugin.Name,
			Env:        env,
			Command:    plugin.Docker.Command,
			Entrypoint: plugin.Docker.Entrypoint,
		}

		if len(spec.Command) == 0 && len(spec.Entrypoint) == 0 {
			spec.Command = []string{fmt.Sprintf("protoc-gen-%s", plugin.Name)}
		}

		if str, ok := plugin.Opt.(string); ok {
			spec.Parameters = []string{str}
		} else if str, ok := plugin.Opt.([]interface{}); ok {
			spec.Parameters = make([]string, len(str))
			for i, v := range str {
				str, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("unsupported opt type: %T", v)
				}
				spec.Parameters[i] = str
			}
		} else if plugin.Opt != nil {
			return nil, fmt.Errorf("unsupported opt type: %T", plugin.Opt)
		}

		out = append(out, spec)
	}

	return out, nil

}

type DockerSpec struct {
	// Optional, defaults to $image/%entrypoint/%cmd
	Name string

	// Only pulls when set
	Pull *PullConfig

	Image      string
	Env        []string
	Entrypoint []string
	Command    []string

	Parameters []string
}

type PullConfig struct {
	Username           string
	PasswordEnvVarName string
}

type BuildSpec struct {
	GoModFile     []byte
	CommitTime    time.Time
	CommitHash    string
	CommitAliases []string

	Builders []DockerSpec
}

type Uploader interface {
	UploadGoModule(ctx context.Context, version gomodproxy.FullInfo, goModData []byte, zipFile io.ReadCloser) error
}

func BuildImage(ctx context.Context, spec BuildSpec, sourceProto *pluginpb.CodeGeneratorRequest, uploader Uploader) error {

	parsedGoMod, err := modfile.Parse("go.mod", spec.GoModFile, nil)
	if err != nil {
		return err
	}

	if parsedGoMod.Module == nil {
		return fmt.Errorf("no module found in go.mod")
	}

	packageName := parsedGoMod.Module.Mod.Path

	commitHashPrefix := spec.CommitHash
	if len(commitHashPrefix) > 12 {
		commitHashPrefix = commitHashPrefix[:12]
	}

	canonicalVersion := module.PseudoVersion("", "", spec.CommitTime, commitHashPrefix)

	dest, err := os.MkdirTemp("", "docker")
	if err != nil {
		return err
	}

	defer os.RemoveAll(dest)

	if err := os.WriteFile(filepath.Join(dest, "go.mod"), spec.GoModFile, 0644); err != nil {
		return err
	}

	pulledImages := map[string]bool{}

	packageRoot := filepath.Join(dest, "package")
	for _, builder := range spec.Builders {

		if builder.Name == "" {
			builder.Name = strings.Join([]string{
				builder.Image, strings.Join(builder.Entrypoint, ","), strings.Join(builder.Command, ","),
			}, "/")
		}

		ctx := log.WithField(ctx, "builder", builder.Name)
		log.Info(ctx, "running build")

		parameter := strings.Join(builder.Parameters, ",")
		sourceProto.Parameter = &parameter

		reqBytes, err := proto.Marshal(sourceProto)
		if err != nil {
			return err
		}

		resp := pluginpb.CodeGeneratorResponse{}

		pull := false
		if builder.Pull != nil && !pulledImages[builder.Image] {
			pull = true
			pulledImages[builder.Image] = true
		}

		outBuffer := &bytes.Buffer{}
		err = DockerRun(ctx, builder, pull, func(hj types.HijackedResponse) error {
			if _, err := hj.Conn.Write(reqBytes); err != nil {
				return err
			}
			if err := hj.CloseWrite(); err != nil {
				return err
			}

			_, err = stdcopy.StdCopy(outBuffer, os.Stderr, hj.Reader)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("running docker %s (%v) %w", builder.Image, builder.Command, err)
		}

		if err := proto.Unmarshal(outBuffer.Bytes(), &resp); err != nil {
			return err
		}

		for _, f := range resp.File {
			name := f.GetName()
			fullPath := filepath.Join(packageRoot, name)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(fullPath, []byte(f.GetContent()), 0644); err != nil {
				return err
			}
		}

		log.Info(ctx, "build complete")
	}

	zipFilePath := filepath.Join(dest, canonicalVersion+".zip")

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

	info := gomodproxy.FullInfo{
		Version:            canonicalVersion,
		VersionAliases:     spec.CommitAliases,
		Time:               spec.CommitTime,
		OriginalCommitHash: spec.CommitHash,
		Package:            packageName,
	}

	zipReader, err := os.Open(zipFilePath)
	if err != nil {
		return err
	}

	return uploader.UploadGoModule(ctx, info, spec.GoModFile, zipReader)

}

func DockerRun(ctx context.Context, spec DockerSpec, pull bool, callback func(hj types.HijackedResponse) error) error {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	if pull && spec.Pull != nil {

		log.Info(ctx, "pulling image")

		username := spec.Pull.Username
		password := os.Getenv(spec.Pull.PasswordEnvVarName)
		cred, _ := json.Marshal(struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: username,
			Password: password,
		})
		reader, err := cli.ImagePull(ctx, spec.Image, types.ImagePullOptions{
			RegistryAuth: string(cred),
		})
		if err != nil {
			return err
		}

		// cli.ImagePull is asynchronous.
		// The reader needs to be read completely for the pull operation to complete.
		// If stdout is not required, consider using io.Discard instead of os.Stdout.
		_, err = io.Copy(os.Stdout, reader)
		reader.Close()
		if err != nil {
			return err
		}
	}

	log.Info(ctx, "creating container")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
		OpenStdin:    true,

		Tty: false,

		Image:      spec.Image,
		Env:        spec.Env,
		Entrypoint: spec.Entrypoint,
		Cmd:        spec.Command,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	defer func() {
		log.Info(ctx, "removing container")
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
			log.WithError(ctx, err).Error("failed to remove container")
		}
	}()

	log.Info(ctx, "attaching to container")

	hj, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
		Logs:   true,
	})
	if err != nil {
		return err
	}

	defer hj.Close()

	log.Info(ctx, "starting container")

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	log.Info(ctx, "run pipe callback")

	if err := callback(hj); err != nil {
		return err
	}

	log.Info(ctx, "waiting for container")

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case st := <-statusCh:
		if st.StatusCode != 0 {
			return fmt.Errorf("non-zero exit code: %d", st.StatusCode)
		}

	}

	log.Info(ctx, "docker build done")

	return nil
}
