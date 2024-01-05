package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pentops/registry/gomodproxy"
	"github.com/pentops/runner/commander"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var Version = "0.0.0"

func main() {
	cmdGroup := commander.NewCommandSet()

	gomodGroup := commander.NewCommandSet()
	gomodGroup.Add("serve", commander.NewCommand(runProxyServer))
	gomodGroup.Add("upload", commander.NewCommand(gomodproxy.RunUploadCommand))
	cmdGroup.Add("gomod", gomodGroup)

	cmdGroup.RunMain("mod-proxy", Version)
}

func runProxyServer(ctx context.Context, cfg struct {
	Port int    `env:"PORT" default:"8080"`
	Src  string `env:"GOMOD_REGISTRY"`
}) error {

	var data gomodproxy.ModProvider
	var err error

	if strings.HasPrefix(cfg.Src, "s3://") {

		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		s3Client := s3.NewFromConfig(awsConfig)

		data, err = gomodproxy.NewS3PackageSrc(ctx, s3Client, cfg.Src)
		if err != nil {
			return err
		}
	} else {
		root := os.DirFS(cfg.Src)
		data, err = gomodproxy.NewLocalPackageMap(ctx, root)
		if err != nil {
			return err
		}
	}

	return gomodproxy.Serve(ctx, cfg.Port, data)
}
