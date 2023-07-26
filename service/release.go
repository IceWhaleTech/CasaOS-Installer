package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	lastRelease        *codegen.Release
	ErrReleaseNotFound = fmt.Errorf("release not found")
)

func GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	var release *codegen.Release
	var mirror string
	for _, mirror = range config.ServerInfo.Mirrors {
		releaseURL := fmt.Sprintf("%s/get/%s/casaos-release", strings.TrimSuffix(mirror, "/"), tag)

		logger.Info("trying to get release information from url", zap.String("url", releaseURL))

		_release, err := internal.GetReleaseFrom(ctx, releaseURL)
		if err != nil {
			logger.Info("error while getting release information - skipping", zap.Error(err), zap.String("url", releaseURL))
			continue
		}

		release = _release
		break
	}

	if release == nil {
		release = lastRelease
	}

	if release == nil {
		return nil, ErrReleaseNotFound
	}

	return release, nil
}

// returns releaseFilePath if successful
func DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	// check and verify existing packages
	if !force {
		if packageFilePath, err := VerifyRelease(release); err != nil {
			logger.Info("error while verifying release - continue to download", zap.Error(err))
		} else {
			logger.Info("package already exists - skipping")
			return packageFilePath, nil
		}
	}

	if release.Mirrors == nil {
		return "", fmt.Errorf("no mirror found")
	}

	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	var packageFilePath string
	var mirror string

	for _, mirror = range release.Mirrors {
		// download packages if any of them is missing
		{
			packageURL, err := internal.GetPackageURLByCurrentArch(release, mirror)
			if err != nil {
				logger.Info("error while getting package url - skipping", zap.Error(err), zap.Any("release", release))
				continue
			}

			packageFilePath, err = internal.Download(ctx, releaseDir, packageURL)
			if err != nil {
				logger.Info("error while downloading and extracting package - skipping", zap.Error(err), zap.String("package_url", packageURL))
				continue
			}
		}

		// download checksums.txt if it's missing
		{
			checksumsURL := internal.GetChecksumsURL(release, mirror)
			if _, err := internal.Download(ctx, releaseDir, checksumsURL); err != nil {
				logger.Info("error while downloading checksums - skipping", zap.Error(err), zap.String("checksums_url", checksumsURL))
				continue
			}
		}
		break
	}

	if packageFilePath == "" {
		return "", fmt.Errorf("package could not be found - there must be a bug")
	}

	release.Mirrors = []string{mirror}

	buf, err := yaml.Marshal(release)
	if err != nil {
		return "", err
	}

	releaseFilePath := filepath.Join(releaseDir, common.ReleaseYAMLFileName)

	return releaseFilePath, os.WriteFile(releaseFilePath, buf, 0o600)
}

func ExtractReleasePackages(packageFilepath string, release codegen.Release) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	if err := internal.Extract(packageFilepath, releaseDir); err != nil {
		return err
	}

	return internal.BulkExtract(releaseDir)
}

func ShouldUpgrade(release codegen.Release) bool {
	if release.Version == "" {
		return false
	}

	targetVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		logger.Info("error while parsing target release version - considered as not upgradable", zap.Error(err), zap.String("release_version", release.Version))
		return false
	}

	currentVersion, err := CurrentReleaseVersion()
	if err != nil {
		logger.Info("error while getting current release version - considered as not upgradable", zap.Error(err))
		return false
	}

	if !targetVersion.GreaterThan(currentVersion) {
		return false
	}

	return true
}

// to check the new version is upgradable and packages are already cached(download)
func IsUpgradable(release codegen.Release) bool {
	if !ShouldUpgrade(release) {
		return false
	}

	_, err := VerifyRelease(release)
	return err == nil
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

	if !VerifyUninstallScript() {
		return fmt.Errorf("uninstall script is not installed")
	}

	return nil
}

func VerifyRelease(release codegen.Release) (string, error) {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	checksums, err := GetChecksums(release)
	if err != nil {
		return "", err
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)
	packageChecksum := checksums[packageFilename]

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	return packageFilePath, VerifyChecksumByFilePath(packageFilePath, packageChecksum)
}

func VerifyUninstallScript() bool {
	// to check the present of file
	// how to do the test? the uninstall is always in the same place?
	return file.CheckNotExist("/usr/bin/casaos-uninstall")
}
