package gomodproxy

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"time"
)

type Info struct {
	Version string    // version string
	Time    time.Time // commit time
}

type VersionNotFoundError string

func (vnfe VersionNotFoundError) Error() string {
	return fmt.Sprintf("version not found: %s", string(vnfe))
}

var NotImplementedError = fmt.Errorf("not implemented")

type PackageProvider interface {
	Info(ctx context.Context, version string) (*Info, error)
	Latest(ctx context.Context) (*Info, error)
	Mod(ctx context.Context, version string) ([]byte, error)
	Zip(ctx context.Context, version string) (io.ReadCloser, error)
	List(ctx context.Context) ([]string, error)
}

type PackageMap map[string]PackageProvider

func (pm PackageMap) Latest(ctx context.Context, packageName string) (*Info, error) {
	pkg, ok := pm[packageName]
	if !ok {
		return nil, VersionNotFoundError(packageName)
	}
	return pkg.Latest(ctx)
}

func (pm PackageMap) List(ctx context.Context, packageName string) ([]string, error) {
	pkg, ok := pm[packageName]
	if !ok {
		return nil, VersionNotFoundError(packageName)
	}
	versions, err := pkg.List(ctx)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (pm PackageMap) Info(ctx context.Context, packageName, version string) (*Info, error) {
	pkg, ok := pm[packageName]
	if !ok {
		return nil, VersionNotFoundError(packageName)
	}
	return pkg.Info(ctx, version)
}

func (pm PackageMap) Mod(ctx context.Context, packageName, version string) ([]byte, error) {
	pkg, ok := pm[packageName]
	if !ok {
		return nil, VersionNotFoundError(packageName)
	}
	return pkg.Mod(ctx, version)
}

func (pm PackageMap) Zip(ctx context.Context, packageName, version string) (io.ReadCloser, error) {
	pkg, ok := pm[packageName]
	if !ok {
		return nil, VersionNotFoundError(packageName)
	}
	return pkg.Zip(ctx, version)
}

func (pm PackageMap) AddVersion(info Info, packageName string, root fs.FS) error {
	anyPkg, ok := pm[packageName]
	if !ok {
		pkg := &LocalPackage{
			Canonical: info.Version,
			Versions:  map[string]Info{},
			Alias:     map[string]string{},
			Root:      root,
		}
		anyPkg = pkg
		pm[packageName] = anyPkg
	}

	pkg, ok := anyPkg.(*LocalPackage)
	if !ok {
		return fmt.Errorf("package %s already exists and is not a local package", info.Version)
	}

	pkg.Versions[info.Version] = info
	if pkg.LatestVersion == "" || info.Time.After(pkg.Versions[pkg.LatestVersion].Time) {
		pkg.LatestVersion = info.Version
	}

	return nil
}

func (pm PackageMap) AddAlias(name string, alias string, root fs.FS) error {
	anyPkg, ok := pm[name]
	if !ok {
		pkg := &LocalPackage{
			Canonical: name,
			Versions:  map[string]Info{},
			Alias:     map[string]string{},
			Root:      root,
		}
		anyPkg = pkg
		pm[name] = anyPkg
	}

	pkg, ok := anyPkg.(*LocalPackage)
	if !ok {
		return fmt.Errorf("package %s already exists and is not a local package", name)
	}

	pkg.Alias[alias] = pkg.Canonical

	return nil

}
