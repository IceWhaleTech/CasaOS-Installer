package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	Cache              = cache.New(5*time.Minute, 10*time.Minute)
	ErrReleaseNotFound = fmt.Errorf("release not found")
)

func GetRelease(tag string) (*codegen.Release, error) {
	cacheKeyPrefix := "release_"

	if cached, ok := Cache.Get(cacheKeyPrefix + tag); ok {
		if release, ok := cached.(*codegen.Release); ok {
			return release, nil
		}
	}

	var release *codegen.Release
	for _, baseURL := range config.ServerInfo.Mirrors {
		releaseURL := fmt.Sprintf("%s/%s/casaos-release", strings.TrimSuffix(baseURL, "/"), tag)

		logger.Info("trying to get release information from url", zap.String("url", releaseURL))

		_release, err := internal.GetReleaseFrom(releaseURL)
		if err != nil {
			logger.Info("error while getting release information - skipping", zap.Error(err), zap.String("url", releaseURL))
			continue
		}

		release = _release
		break
	}

	if release == nil {
		return nil, ErrReleaseNotFound
	}

	Cache.Set(cacheKeyPrefix+tag, release, cache.DefaultExpiration)

	return release, nil
}

func DownloadRelease(ctx context.Context, release codegen.Release) (string, error) {
	if release.Mirrors == nil {
		return "", fmt.Errorf("no mirror found")
	}

	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	var mirror string

	for _, mirror = range release.Mirrors {
		packageURL, err := internal.GetPackageURLByCurrentArch(release, mirror)
		if err != nil {
			logger.Info("error while getting package url - skipping", zap.Error(err), zap.String("mirror", mirror))
			continue
		}

		if err := internal.DownloadAndExtractPackage(ctx, releaseDir, packageURL); err != nil {
			logger.Info("error while downloading and extracting package - skipping", zap.Error(err), zap.String("package_url", packageURL))
			continue
		}
		break
	}

	release.Mirrors = []string{mirror}

	buf, err := yaml.Marshal(release)
	if err != nil {
		return "", err
	}

	releaseFilePath := filepath.Join(releaseDir, common.ReleaseYAMLFilename)

	return releaseFilePath, os.WriteFile(releaseFilePath, buf, 0o600)
}

func InstallRelease(ctx context.Context, release codegen.Release, sysrootPath string, tryDownload bool) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	releaseFilePath := filepath.Join(releaseDir, common.ReleaseYAMLFilename)
	if _, err := os.Stat(releaseFilePath); os.IsNotExist(err) && tryDownload {
		logger.Info("release file not found - downloading...", zap.String("release_file_path", releaseFilePath))
		if _, err := DownloadRelease(ctx, release); err != nil {
			return err
		}
	}

	backgroundCtx := context.Background()
	if err := internal.InstallRelease(backgroundCtx, releaseDir, sysrootPath); err != nil {
		return err
	}

	return nil
}

func ReleaseDir(release codegen.Release) (string, error) {
	if release.Version == "" {
		return "", fmt.Errorf("release version is empty")
	}

	return filepath.Join(config.ServerInfo.CachePath, "releases", release.Version), nil
}
