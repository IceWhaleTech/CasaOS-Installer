package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Common/utils/systemctl"
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

func DownloadUninstallScript(ctx context.Context, sysRoot string) (string, error) {
	CASA_UNINSTALL_URL := "https://get.casaos.io/uninstall/v0.4.0"
	CASA_UNINSTALL_PATH := filepath.Join(sysRoot, "/usr/bin/casaos-uninstall")
	// to delete the old uninstall script when the script is exsit
	if _, err := os.Stat(CASA_UNINSTALL_PATH); err == nil {
		// 删除文件
		err := os.Remove(CASA_UNINSTALL_PATH)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Old uninstall script deleted successfully")
		}
	}

	// to download the new uninstall script
	if err := internal.DownloadAs(ctx, CASA_UNINSTALL_PATH, CASA_UNINSTALL_URL); err != nil {
		return CASA_UNINSTALL_PATH, err
	}
	// change the permission of the uninstall script
	if err := os.Chmod(CASA_UNINSTALL_PATH, 0o755); err != nil {
		return CASA_UNINSTALL_PATH, err
	}

	return "", nil
}

// returns releaseFilePath if successful
func DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {

	go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateBegin, nil)
	defer PublishEventWrapper(ctx, common.EventTypeDownloadUpdateEnd, nil)

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
		go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
			common.PropertyTypeMessage.Name: "no mirror found",
		})
		return "", fmt.Errorf("no mirror found")
	}

	releaseDir, err := ReleaseDir(release)
	if err != nil {
		go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
			common.PropertyTypeMessage.Name: err.Error(),
		})
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
				go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
					common.PropertyTypeMessage.Name: err.Error(),
				})
				continue
			}

			packageFilePath, err = internal.Download(ctx, releaseDir, packageURL)
			if err != nil {
				logger.Info("error while downloading and extracting package - skipping", zap.Error(err), zap.String("package_url", packageURL))
				go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
					common.PropertyTypeMessage.Name: err.Error(),
				})
				continue
			}
		}

		// download checksums.txt if it's missing
		{
			checksumsURL := internal.GetChecksumsURL(release, mirror)
			if _, err := internal.Download(ctx, releaseDir, checksumsURL); err != nil {
				logger.Info("error while downloading checksums - skipping", zap.Error(err), zap.String("checksums_url", checksumsURL))
				go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
					common.PropertyTypeMessage.Name: err.Error(),
				})
				continue
			}
		}
		break
	}

	if packageFilePath == "" {
		go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateError, map[string]string{
			common.PropertyTypeMessage.Name: "package could not be found - there must be a bug",
		})
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

func ShouldUpgrade(release codegen.Release, sysrootPath string) bool {
	if release.Version == "" {
		return false
	}

	targetVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		logger.Info("error while parsing target release version - considered as not upgradable", zap.Error(err), zap.String("release_version", release.Version))
		return false
	}

	currentVersion, err := CurrentReleaseVersion(sysrootPath)
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
func IsUpgradable(release codegen.Release, sysrootPath string) bool {
	if !ShouldUpgrade(release, sysrootPath) {
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

	return nil
}
func InstallDependencies(ctx context.Context, release codegen.Release, sysrootPath string) error {
	internal.InstallDependencies()
	return nil
}

func PostReleaseInstall(ctx context.Context, release codegen.Release, sysrootPath string) error {
	// post release install script
	// work list
	// 1. overwrite target release
	targetReleaseLocalPath = filepath.Join(sysrootPath, targetReleaseLocalPath)
	targetReleaseContent, err := yaml.Marshal(release)
	if err != nil {
		return err
	}
	if err := os.WriteFile(targetReleaseLocalPath, targetReleaseContent, 0o666); err != nil {
		return err
	}

	// 2. if current release is not exist, create it( using current release version )
	// if current release is exist, It mean the casaos is old casaos that install by shell
	// So It should update to casaos v0.4.4 and we didn't need to migrate it.
	currentReleaseLocalPath = filepath.Join(sysrootPath, currentReleaseLocalPath)
	if _, err := os.Stat(currentReleaseLocalPath); os.IsNotExist(err) {
		currentReleaseContent, err := yaml.Marshal(release)
		if err != nil {
			return err
		}
		if err := os.WriteFile(currentReleaseLocalPath, currentReleaseContent, 0o666); err != nil {
			return err
		}
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
	return !file.CheckNotExist("/usr/bin/casaos-uninstall")
}

func ExecuteModuleInstallScript(releaseFilePath string, release codegen.Release) error {
	// run setup script
	scriptFolderPath := filepath.Join(releaseFilePath, "..", "build/scripts/setup/script.d")
	// to get the script file name from scriptFolderPath
	// to execute the script in name order
	filepath.WalkDir(scriptFolderPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		cmd := exec.Command(path)
		err = cmd.Run()
		return err
	})

	// // run service script
	// serviceScriptFolderPath := filepath.Join(releaseFilePath, "..", "build/scripts/setup/service.d")
	// for _, module := range release.Modules {
	// 	moduleServiceScriptFolderPath := filepath.Join(serviceScriptFolderPath, module.Name)
	// }

	return nil
}

func enableAndStartSystemdService(serviceName string) error {
	// if err := systemctl.EnableService(fmt.Sprintf("%s.service", serviceName)); err != nil {
	// 	return err
	// }
	if err := systemctl.StartService(fmt.Sprintf("%s.service", serviceName)); err != nil {
		return err
	}
	return nil
}
func SetStartUpAndLaunchModule(release codegen.Release) error {
	for _, module := range release.Modules {
		if err := enableAndStartSystemdService(module.Name); err != nil {
			return err
		}
	}
	return nil
}
