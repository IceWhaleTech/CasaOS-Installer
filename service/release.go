package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

var (
	CurrentReleaseLocalPath = "/etc/release.yaml"
)

type InstallerType string

const (
	RAUC        InstallerType = "rauc"
	RAUCOFFLINE InstallerType = "rauc_offline"
	TAR         InstallerType = "tar"
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
	if release.Mirrors == nil {
		return "", fmt.Errorf("no mirror found")
	}

	releaseDir, err := config.ReleaseDir(release)
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

func IsZimaOS(sysRoot string) bool {
	// read sysRoot/etc/os-release
	// if the file have "MODEL="Zima" return true
	// else return false
	fileContent, err := os.ReadFile(filepath.Join(sysRoot, "etc/os-release"))
	if err != nil {
		return false
	}
	if strings.Contains(string(fileContent), "MODEL=Zima") {
		return true
	}
	return false
}

func IsCasaOS(sysRoot string) bool {
	fileContent, err := os.ReadFile(filepath.Join(sysRoot, "etc/os-release"))
	if err != nil {
		return true
	}
	if strings.Contains(string(fileContent), "MODEL=Zima") {
		return false
	}
	return true
}

func GetReleaseBranch(sysRoot string) string {
	// return "rauc"

	if IsZimaOS(sysRoot) {
		return "rauc"
	}
	if IsCasaOS(sysRoot) {
		return "dev-test"
	}
	return "main"
}

func GetInstallMethod(sysRoot string) (InstallerType, error) {
	// to check the system is casaos or zimaos
	// if zimaos, return "rauc"
	// if casaos, return "tar"

	if IsZimaOS(sysRoot) {
		// to check file exsit
		if _, err := os.Stat(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)); os.IsNotExist(err) {
			return RAUC, nil
		} else {
			return RAUCOFFLINE, nil
		}
	}
	if IsCasaOS(sysRoot) {
		return TAR, nil
	}
	return "", fmt.Errorf("unknown system")
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

func InstallRelease(release codegen.Release, sysrootPath string) error {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return err
	}

	if err := internal.InstallRelease(releaseDir, sysrootPath); err != nil {
		return err
	}

	return nil
}

func InstallDependencies(release codegen.Release, sysrootPath string) error {
	internal.InstallDependencies()
	return nil
}
func VerifyRelease(release codegen.Release) (string, error) {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)

	// 回头再把这个开一下
	// packageChecksum := checksums[packageFilename]

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	// TODO 以后做一个优化，以后tar用tar的verify，分开实现
	// 当前我都不做校验
	// return packageFilePath, VerifyChecksumByFilePath(packageFilePath, packageChecksum)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return packageFilePath, err
	}
	return packageFilePath, nil
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

		fmt.Println("执行: ", path)
		cmd := exec.Command(path)
		err = cmd.Run()
		return err
	})

	return nil
}

func reStartSystemdService(serviceName string) error {
	// TODO remove the code, because the service is stop in before
	// but in install rauc. the stop is important. So I need to think about it.
	if err := systemctl.StopService(fmt.Sprintf("%s.service", serviceName)); err != nil {
		return err
	}

	if err := systemctl.StartService(fmt.Sprintf("%s.service", serviceName)); err != nil {
		return err
	}
	return nil
}

func stopSystemdService(serviceName string) error {
	if err := systemctl.StopService(fmt.Sprintf("%s.service", serviceName)); err != nil {
		return err
	}
	return nil

}

func StopModule(release codegen.Release) error {
	err := error(nil)
	for _, module := range release.Modules {
		fmt.Println("停止: ", module.Name)
		if err := stopSystemdService(module.Name); err != nil {
			fmt.Printf("failed to stop module: %s\n", err.Error())
		}
		// to sleep 1s
		time.Sleep(1 * time.Second)
	}
	return err
}

func LaunchModule(release codegen.Release) error {
	for _, module := range release.Modules {
		fmt.Println("启动: ", module.Name)
		if err := reStartSystemdService(module.Name); err != nil {
			return err
		}
		// to sleep 1s
		time.Sleep(1 * time.Second)
	}
	return nil
}
