package gomodproxy

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

func NewLocalPackageMap(ctx context.Context, root fs.FS) (PackageMap, error) {

	var data = PackageMap{}
	if err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		packageName := filepath.Dir(path)
		sub, err := fs.Sub(root, packageName)
		if err != nil {
			return err
		}

		switch filepath.Ext(path) {
		case ".info":
			infoFile, err := root.Open(path)
			if err != nil {
				return err
			}
			defer infoFile.Close()

			info := Info{}
			if err := json.NewDecoder(infoFile).Decode(&info); err != nil {
				return err
			}

			if err := data.AddVersion(info, packageName, sub); err != nil {
				return err
			}

		case ".alias":
			aliasFile, err := root.Open(path)
			if err != nil {
				return err
			}
			defer aliasFile.Close()

			aliasDestBytes, err := io.ReadAll(aliasFile)
			if err != nil {
				return err
			}

			aliasDest := strings.TrimSpace(string(aliasDestBytes))

			if err := data.AddAlias(packageName, aliasDest, sub); err != nil {
				return err
			}

		}

		return nil

	}); err != nil {
		return nil, err
	}

	return data, nil
}

type LocalPackage struct {
	Canonical     string
	Versions      map[string]Info
	LatestVersion string
	Alias         map[string]string
	Root          fs.FS
}

func (lps LocalPackage) Info(ctx context.Context, version string) (*Info, error) {
	return lps.resolveVersion(version)
}

func (pkg LocalPackage) Latest(ctx context.Context) (*Info, error) {
	latest, ok := pkg.Versions[pkg.LatestVersion]
	if !ok {
		return nil, VersionNotFoundError(pkg.LatestVersion)
	}
	return &latest, nil
}

func (pkg LocalPackage) Mod(ctx context.Context, version string) ([]byte, error) {
	info, err := pkg.resolveVersion(version)
	if err != nil {
		return nil, err
	}

	file, err := pkg.Root.Open(info.Version + ".mod")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (pkg LocalPackage) Zip(ctx context.Context, version string) (io.ReadCloser, error) {
	info, err := pkg.resolveVersion(version)
	if err != nil {
		return nil, err
	}

	file, err := pkg.Root.Open(info.Version + ".zip")
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (pkg LocalPackage) resolveVersion(given string) (*Info, error) {
	if info, ok := pkg.Versions[given]; ok {
		return &info, nil
	}
	if v, ok := pkg.Alias[given]; ok {
		info, ok := pkg.Versions[v]
		if !ok {
			return nil, VersionNotFoundError(v)
		}
		return &info, nil
	}
	return nil, VersionNotFoundError(given)
}

func (pkg LocalPackage) List(ctx context.Context) ([]string, error) {
	versions := make([]string, 0, len(pkg.Versions))
	for _, v := range pkg.Versions {
		versions = append(versions, v.Version)
	}
	return versions, nil
}
