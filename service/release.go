package service

import (
	"context"
	"fmt"
	"net/http"
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

	"github.com/samber/lo"
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

type ConstructReleaseFileUrlFunc func(tag string, mirror string) string

func GitHubBranchTagReleaseUrl(tag string, _ string) string {
	// 这个不走新的mirror，自己用的旧的，回头迁移完，这个函数就删了。
	mirror := "https://raw.githubusercontent.com/IceWhaleTech"
	return fmt.Sprintf("%s/get/%s/casaos-release", strings.TrimSuffix(mirror, "/"), tag)
}

func HyperFileTagReleaseUrl(tag string, mirror string) string {
	// https://raw.githubusercontent.com/IceWhaleTech/zimaos-rauc/main/rauc
	// https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/rauc
	return mirror + tag
}

type BestURLFunc func(urls []string) string

func BestByDelay(urls []string) string {
	type result struct {
		url     string
		latency time.Duration
	}

	ch := make(chan result)

	for _, url := range urls {
		go func(url string) {
			start := time.Now()
			resp, err := http.Get(url)
			if err != nil || resp.StatusCode != http.StatusOK {
				ch <- result{url: url, latency: 0}
				return
			}
			latency := time.Since(start)
			ch <- result{url: url, latency: latency}
		}(url)
	}

	var first result
	for range urls {
		res := <-ch
		if res.latency != 0 && (first.latency == 0 || res.latency < first.latency) {
			first = res
		}
	}

	return first.url
}
func FetchRelease(ctx context.Context, tag string, constructReleaseFileUrlFunc ConstructReleaseFileUrlFunc) (*codegen.Release, error) {
	var releaseURL []string

	for _, mirror := range config.ServerInfo.Mirrors {
		releaseURL = append(releaseURL, constructReleaseFileUrlFunc(tag, mirror))
	}

	var best BestURLFunc = BestByDelay // dependency inject

	url := best(releaseURL)
	var release *codegen.Release
	release, err := internal.GetReleaseFrom(ctx, url)
	if err != nil {
		logger.Info("trying to get release information from url", zap.String("url", url))
		return release, err
	}
	return release, nil
}

func GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	var release *codegen.Release
	var mirror string

	releaseURL := GitHubBranchTagReleaseUrl(tag, mirror)

	logger.Info("trying to get release information from url", zap.String("url", releaseURL))

	_release, err := internal.GetReleaseFrom(ctx, releaseURL)
	if err != nil {
		logger.Info("error while getting release information - skipping", zap.Error(err), zap.String("url", releaseURL))
	}

	release = _release

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
		return "rauc"
	}
	return "main"
}

func CheckOfflineTarExist(sysRoot string) bool {
	// get all file from /DATA/rauc
	// if the file have "*.tar" return true
	files := internal.GetAllFile(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH))
	println("files : ", files)

	// only allow one tar file
	tar_files := lo.FilterMap(files, func(filename string, _ int) (string, bool) {
		if strings.HasSuffix(filename, ".tar") {
			return filename, true
		}
		return "", false
	})

	if len(tar_files) == 1 {
		file_name := files[0]
		if strings.HasSuffix(file_name, ".tar") {
			println("find offline rauc file: ", file_name)
			config.RAUC_OFFLINE_RAUC_FILENAME = file_name
			return true
		}
	} else {
		return false
	}
	return false
}

func CheckOfflineRAUCExist(sysRoot string) bool {
	// get all file from /DATA/rauc
	// if the file have "*.tar" return true
	files := internal.GetAllFile(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH))
	println("files : ", files)

	// only allow one tar file
	tar_files := lo.FilterMap(files, func(filename string, _ int) (string, bool) {
		if strings.HasSuffix(filename, ".tar") {
			return filename, true
		}
		return "", false
	})

	if len(tar_files) == 1 {
		file_name := files[0]
		if strings.HasSuffix(file_name, ".tar") {
			println("find offline rauc file: ", file_name)
			config.RAUC_OFFLINE_RAUC_FILENAME = file_name
			return true
		}
	} else {
		return false
	}
	return false
}

func GetInstallMethod(sysRoot string) (InstallerType, error) {
	// to check the system is casaos or zimaos
	// if zimaos, return "rauc"
	// if casaos, return "tar"

	// for test. always open rauc
	if IsZimaOS(sysRoot) || true {
		// to check file exist
		if !CheckOfflineRAUCExist(sysRoot) {
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

func ShouldUpgrade(release codegen.Release, sysRootPath string) bool {
	if release.Version == "" {
		return false
	}

	targetVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		logger.Info("error while parsing target release version - considered as not upgradable", zap.Error(err), zap.String("release_version", release.Version))
		return false
	}

	currentVersion, err := CurrentReleaseVersion(sysRootPath)
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
func IsUpgradable(release codegen.Release, sysRootPath string) bool {
	if !ShouldUpgrade(release, sysRootPath) {
		return false
	}

	_, err := VerifyRelease(release)
	return err == nil
}

func InstallRelease(release codegen.Release, sysRootPath string) error {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return err
	}

	if err := internal.InstallRelease(releaseDir, sysRootPath); err != nil {
		return err
	}

	return nil
}

func InstallDependencies(release codegen.Release, sysRootPath string) error {
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
