package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/pentops/log.go/log"
	"github.com/pentops/registry/builder"
	"golang.org/x/oauth2"
	"gopkg.daemonl.com/envconf"

	"github.com/google/go-github/v58/github"
)

type Client struct {
	repositories RepositoriesService
	git          GitService
}

type GitService interface {
	ListMatchingRefs(ctx context.Context, owner, repo string, opts *github.ReferenceListOptions) ([]*github.Reference, *github.Response, error)
}
type RepositoriesService interface {
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	ListByOrg(context.Context, string, *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	GetArchiveLink(ctx context.Context, owner string, repo string, archiveFormat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
	GetCommit(ctx context.Context, owner string, repo string, ref string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
}

func NewEnvClient(ctx context.Context) (*Client, error) {

	config := struct {
		// Method 1
		GithubToken string `env:"GITHUB_TOKEN" default:""`

		// Method 2
		GithubPrivateKey     string `env:"GH_PRIVATE_KEY" default:""`
		GithubAppID          int64  `env:"GH_APP_ID" default:"0"`
		GithubInstallationID int64  `env:"GH_INSTALLATION_ID" default:"0"`
	}{}

	if err := envconf.Parse(&config); err != nil {
		return nil, err
	}

	var err error
	var client *Client

	if config.GithubPrivateKey != "" {

		if config.GithubAppID == 0 || config.GithubInstallationID == 0 {
			return nil, fmt.Errorf("no github app id or installation id")
		}
		tr := http.DefaultTransport
		privateKey, err := base64.StdEncoding.DecodeString(config.GithubPrivateKey)
		if err != nil {
			return nil, err
		}

		itr, err := ghinstallation.New(tr, config.GithubAppID, int64(config.GithubInstallationID), privateKey)
		if err != nil {
			return nil, err
		}

		client, err = NewClient(&http.Client{Transport: itr})
		if err != nil {
			return nil, err
		}

	} else if config.GithubToken != "" {

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GithubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client, err = NewClient(tc)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no valid github config in environment")
	}

	return client, nil
}

func NewClient(tc *http.Client) (*Client, error) {
	ghcl := github.NewClient(tc)
	cl := &Client{
		repositories: ghcl.Repositories,
		git:          ghcl.Git,
	}
	return cl, nil
}

func (cl Client) GetCommit(ctx context.Context, org string, repo string, ref string) (*builder.CommitInfo, error) {

	commit, _, err := cl.repositories.GetCommit(ctx, org, repo, ref, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	ts := commit.GetCommit().GetCommitter().GetDate()
	info := &builder.CommitInfo{
		Hash: commit.GetSHA(),
		Time: ts.Time,
	}

	refs, _, err := cl.git.ListMatchingRefs(ctx, org, repo, &github.ReferenceListOptions{
		Ref: ref,
	})

	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		refName := ref.GetRef()
		info.Aliases = append(info.Aliases, refName)
	}

	return info, nil
}

func (cl Client) GetContent(ctx context.Context, org string, repo string, ref string, destDir string) error {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	linkURL, _, err := cl.repositories.GetArchiveLink(ctx, org, repo, github.Zipball, opts, 5)
	if err != nil {
		return err
	}

	log.WithField(ctx, "url", linkURL.String()).Debug("downloading")

	getRes, err := http.DefaultClient.Get(linkURL.String())
	if err != nil {
		return err
	}

	if getRes.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d", getRes.StatusCode)
	}

	defer getRes.Body.Close()

	// TODO: Use a disk.
	zipBody, err := io.ReadAll(getRes.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBody), int64(len(zipBody)))
	if err != nil {
		return err
	}

	prefix := ""

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		if prefix == "" {
			parts := strings.Split(file.Name, "/")
			prefix = parts[0]
		}

		if !strings.HasPrefix(file.Name, prefix) {
			return fmt.Errorf("invalid file name %q", file.Name)
		}
		destFile := filepath.Join(destDir, file.Name[len(prefix):])
		destDir := filepath.Dir(destFile)

		if err := func() error {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return err
			}

			dest, err := os.Create(destFile)
			if err != nil {
				return err
			}
			defer dest.Close()

			archiveFile, err := file.Open()
			if err != nil {
				return err
			}

			defer archiveFile.Close()

			if _, err := io.Copy(dest, archiveFile); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}

	}

	return nil
}
