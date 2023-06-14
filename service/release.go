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

func DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	if release.Mirrors == nil {
		return "", fmt.Errorf("no mirror found")
	}

	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	var packageFilepath string
	var mirror string

	for _, mirror = range release.Mirrors {

		var checksum map[string]string

		// checksum
		{
			checksumURL := internal.GetChecksumURL(release, mirror)
			checksumFilePath, err := internal.Download(ctx, releaseDir, checksumURL)
			if err != nil {
				logger.Info("error while downloading checksum - skipping", zap.Error(err), zap.String("checksum_url", checksumURL))
				continue
			}

			_checksum, err := internal.GetChecksum(checksumFilePath)
			if err != nil {
				logger.Info("error while getting checksum - skipping", zap.Error(err), zap.String("checksum_file_path", checksumFilePath))
				continue
			}

			checksum = _checksum
		}

		packageURL, err := internal.GetPackageURLByCurrentArch(release, mirror)
		if err != nil {
			logger.Info("error while getting package url - skipping", zap.Error(err), zap.String("mirror", mirror))
			continue
		}

		packageFilename := filepath.Base(packageURL)
		packageChecksum := checksum[packageFilename]

		// check and verify existing packages
		if !force {
			packageFilepath = filepath.Join(releaseDir, packageFilename)

			if err := VerifyChecksum(packageFilepath, packageChecksum); err != nil {
				logger.Info("error while verifying checksum of package already exists - skipping", zap.Error(err), zap.String("package_file_path", packageFilepath))
			} else {
				logger.Info("package already exists - skipping", zap.String("package_file_path", packageFilepath))
				break
			}
		}

		// download packages if any of them is missing
		{
			packageFilepath, err = internal.Download(ctx, releaseDir, packageURL)
			if err != nil {
				logger.Info("error while downloading and extracting package - skipping", zap.Error(err), zap.String("package_url", packageURL))
				continue
			}

			if err := VerifyChecksum(packageFilepath, packageChecksum); err != nil {
				logger.Info("error while verifying checksum of package just downloaded - skipping", zap.Error(err), zap.String("package_file_path", packageFilepath))
				continue
			}
		}
		break
	}

	if packageFilepath == "" {
		return "", fmt.Errorf("package could not be found - there must be a bug")
	}

	// extract the main package, e.g. casaos-amd64-v0.4.4-alpha2.tar.gz
	if err := internal.Extract(packageFilepath, releaseDir); err != nil {
		return "", err
	}

	// extract individual sub-packages from the main package, e.g. linux-amd64-casaos-app-management-v0.4.4-alpha16.tar.gz
	if err := internal.BulkExtract(releaseDir); err != nil {
		return "", err
	}

	release.Mirrors = []string{mirror}

	buf, err := yaml.Marshal(release)
	if err != nil {
		return "", err
	}

	releaseFilePath := filepath.Join(releaseDir, common.ReleaseYAMLFileName)

	return releaseFilePath, os.WriteFile(releaseFilePath, buf, 0o600)
}

func InstallRelease(ctx context.Context, release codegen.Release, sysrootPath string) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	backgroundCtx := context.Background()
	if err := internal.InstallRelease(backgroundCtx, releaseDir, sysrootPath); err != nil {
		return err
	}

	return nil
}
