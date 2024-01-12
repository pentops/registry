package builder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/log.go/log"
)

var DefaultRegistryAuths = []*config_j5pb.DockerRegistryAuth{{
	Registry: "ghcr.io/*",
	Auth: &config_j5pb.DockerRegistryAuth_Github_{
		Github: &config_j5pb.DockerRegistryAuth_Github{},
	},
}, {
	Registry: "*.dkr.ecr.*.amazonaws.com/*",
	Auth: &config_j5pb.DockerRegistryAuth_AwsEcs{
		AwsEcs: &config_j5pb.DockerRegistryAuth_AWSECS{},
	},
}}

type DockerWrapper struct {
	pulledImages map[string]bool
	client       *client.Client
	auth         []*config_j5pb.DockerRegistryAuth
}

type ECRAPI interface {
	GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

func NewDockerWrapper(registryAuth []*config_j5pb.DockerRegistryAuth) (*DockerWrapper, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerWrapper{
		pulledImages: make(map[string]bool),
		client:       cli,
		auth:         registryAuth,
	}, nil
}

func (dw *DockerWrapper) Close() error {
	return dw.client.Close()
}

func (dw *DockerWrapper) Run(ctx context.Context, spec *config_j5pb.DockerSpec, input io.Reader, output io.Writer) error {

	if spec.Pull {
		// skip if pulled...
		if !dw.pulledImages[spec.Image] {
			// only pull once for all plugins
			dw.pulledImages[spec.Image] = true

			log.Info(ctx, "pulling image")
			if err := dw.Pull(ctx, spec); err != nil {
				log.WithError(ctx, err).Error("failed to pull image")
				return err
			}
			log.Info(ctx, "image pulled")
		}
	}

	env, err := mapEnvVars(spec.Env)
	if err != nil {
		return err
	}

	log.Info(ctx, "creating container")

	resp, err := dw.client.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
		OpenStdin:    true,

		Tty: false,

		Image:      spec.Image,
		Env:        env,
		Entrypoint: spec.Entrypoint,
		Cmd:        spec.Command,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	defer func() {
		log.Info(ctx, "removing container")
		if err := dw.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
			log.WithError(ctx, err).Error("failed to remove container")
		}
	}()

	log.Info(ctx, "attaching to container")

	hj, err := dw.client.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
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

	if err := dw.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	log.Info(ctx, "run pipe callback")

	if _, err := io.Copy(hj.Conn, input); err != nil {
		return err
	}
	if err := hj.CloseWrite(); err != nil {
		return err
	}

	_, err = stdcopy.StdCopy(output, os.Stderr, hj.Reader)
	if err != nil {
		return err
	}

	log.Info(ctx, "waiting for container")

	statusCh, errCh := dw.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

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

func mapEnvVars(spec []string) ([]string, error) {
	env := make([]string, len(spec))
	for idx, src := range spec {
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
	return env, nil
}

func globMatch(pattern, s string) bool {
	escaped := regexp.QuoteMeta(pattern)
	// Replace escaped * with .* to make it a regexp pattern.
	pattern = strings.ReplaceAll(escaped, "\\*", ".*")
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return matcher.MatchString(s)
}

func (dw *DockerWrapper) Pull(ctx context.Context, spec *config_j5pb.DockerSpec) error {

	type basicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	pullOptions := types.ImagePullOptions{}

	var registryAuth *config_j5pb.DockerRegistryAuth
	for _, auth := range dw.auth {
		// If auth's registry pattern with * wildcards matches the spec's image, use it.
		if globMatch(auth.Registry, spec.Image) {
			registryAuth = auth
			log.WithField(ctx, "registry", auth.Registry).Debug("using auth")
			break
		}
	}
	if registryAuth == nil {
		log.WithField(ctx, "image", spec.Image).Debug("no registry auth matched")
	}

	if registryAuth != nil {
		var authConfig *basicAuth

		switch authType := registryAuth.Auth.(type) {
		case *config_j5pb.DockerRegistryAuth_Basic_:
			authConfig = &basicAuth{
				Username: authType.Basic.Username,
				Password: os.Getenv(authType.Basic.PasswordEnvVar),
			}

		case *config_j5pb.DockerRegistryAuth_AwsEcs:

			// TODO: This is a little too magic.
			awsConfig, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			ecrClient := ecr.NewFromConfig(awsConfig)
			resp, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
			if err != nil {
				return fmt.Errorf("failed to get authorization token: %w", err)
			}

			if len(resp.AuthorizationData) == 0 {
				return fmt.Errorf("no authorization data returned")
			}

			authData, err := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
			if err != nil {
				return fmt.Errorf("failed to decode authorization token: %w", err)
			}

			parts := strings.SplitN(string(authData), ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid authorization token")
			}

			authConfig = &basicAuth{
				Username: parts[0],
				Password: parts[1],
			}

		default:
			return fmt.Errorf("unknown auth type: %T", authType)
		}
		cred, _ := json.Marshal(authConfig)
		pullOptions.RegistryAuth = base64.StdEncoding.EncodeToString(cred)
	}

	reader, err := dw.client.ImagePull(ctx, spec.Image, pullOptions)
	if err != nil {
		return fmt.Errorf("image pull: %w", err)
	}

	// cli.ImagePull is asynchronous.
	// The reader needs to be read completely for the pull operation to complete.
	// If stdout is not required, consider using io.Discard instead of os.Stdout.
	_, err = io.Copy(os.Stdout, reader)
	reader.Close()
	if err != nil {
		return err
	}
	return nil
}
