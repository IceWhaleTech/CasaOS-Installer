package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
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

var CurrentReleaseLocalPath = "/etc/release.yaml"

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
	return mirror + tag + ".txt"
}

type BestURLFunc func(urls []string) string

func BestByDelay(urls []string) string {
	ch := make(chan string)
	for _, url := range urls {
		go func(url string) {
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			resp, err := client.Head(url)
			if err != nil || resp.StatusCode != http.StatusOK {
				return
			}
			ch <- url
		}(url)
	}

	first := <-ch
	config.ServerInfo.BestUrl = first
	return first
}

func FetchRelease(ctx context.Context, tag string, constructReleaseFileUrlFunc ConstructReleaseFileUrlFunc) (*codegen.Release, error) {
	url := config.ServerInfo.BestUrl
	if len(config.ServerInfo.BestUrl) == 0 {
		var releaseURL []string
		for _, mirror := range config.ServerInfo.Mirrors {
			releaseURL = append(releaseURL, constructReleaseFileUrlFunc(tag, mirror))
		}
		var best BestURLFunc = BestByDelay // dependency inject
		url = best(releaseURL)
	}
	logger.Info("fetching release", zap.String("tag", tag), zap.String("url", url))
	var release *codegen.Release
	release, err := internal.GetReleaseFrom(ctx, url)
	if err != nil {
		logger.Error("failed to get release information from url", zap.String("url", url), zap.Error(err))
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
				logger.Error("error while getting package url - skipping", zap.Error(err), zap.Any("release", release))
				continue
			}

			packageFilePath, err = internal.Download(ctx, releaseDir, packageURL)
			if err != nil {
				logger.Error("error while downloading and extracting package - skipping", zap.Error(err), zap.String("package_url", packageURL))
				continue
			}
			logger.Info("downloaded package success", zap.String("package_url", packageURL), zap.String("package_file_path", packageFilePath))
		}

		// download checksums.txt if it's missing
		{
			checksumsURL := internal.GetChecksumsURL(release, mirror)
			if _, err := internal.Download(ctx, releaseDir, checksumsURL); err != nil {
				logger.Error("error while downloading checksums - skipping", zap.Error(err), zap.String("checksums_url", checksumsURL))
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

func CheckOfflineRAUCExist(sysRoot string) bool {
	// get all file from /DATA/rauc
	// if the file have "*.tar" return true
	files := internal.GetAllFile(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH))

	// only allow one tar file
	raucb_files := lo.FilterMap(files, func(filename string, _ int) (string, bool) {
		if strings.HasSuffix(filename, ".raucb") {
			return filename, true
		}
		return "", false
	})

	if len(raucb_files) >= 1 {
		file_name := raucb_files[0]
		if strings.HasSuffix(file_name, ".raucb") {
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

func parseVersion(v string) (major, minor, patch, revision int, tag string) {
	parts := strings.Split(v, ".")
	major, _ = strconv.Atoi(parts[0])
	minor, _ = strconv.Atoi(parts[1])

	tagParts := strings.Split(parts[2], "-")
	patch, _ = strconv.Atoi(tagParts[0])

	if len(tagParts) > 1 {
		revision, _ = strconv.Atoi(tagParts[1])
		if strings.Contains(tagParts[1], "alpha") {
			tag = "alpha"
			revision, _ = strconv.Atoi(strings.Trim(tagParts[1], "alpha"))
		} else if strings.Contains(tagParts[1], "beta") {
			tag = "beta"
			revision, _ = strconv.Atoi(strings.Trim(tagParts[1], "beta"))
		}
	}

	return
}

// Helper function to compare version tags, returns true if targetTag is considered higher
func compareTags(currentTag, targetTag string) bool {
	// Define order for known tags
	tagPriority := map[string]int{
		"alpha": 1,
		"beta":  2,
	}

	currentTagValue, currentTagExists := tagPriority[currentTag]
	targetTagValue, targetTagExists := tagPriority[targetTag]

	// If both tags are known, compare their priorities
	if currentTagExists && targetTagExists {
		return targetTagValue > currentTagValue
	}

	// If one tag is unknown, it is considered lower than known tags
	return !targetTagExists
}

func IsNewerVersionString(current string, target string) bool {
	currentMajor, currentMinor, currentPatch, currentRevision, currentTag := parseVersion(current)
	targetMajor, targetMinor, targetPatch, targetRevision, targetTag := parseVersion(target)

	// Compare major versions
	if targetMajor != currentMajor {
		return targetMajor > currentMajor
	}
	// Compare minor versions
	if targetMinor != currentMinor {
		return targetMinor > currentMinor
	}
	// Compare patch versions
	if targetPatch != currentPatch {
		return targetPatch > currentPatch
	}

	// Compare tag values where no tag is considered higher than any tagged version
	if currentTag == "" && targetTag != "" {
		return false
	} else if currentTag != "" && targetTag == "" {
		return true
	} else if currentTag != targetTag {
		return compareTags(currentTag, targetTag)
	}

	// Compare revisions under the same tag or revision number if no tag exists
	return targetRevision > currentRevision
}

func IsNewerVersion(current *semver.Version, target *semver.Version) bool {
	return IsNewerVersionString(current.String(), target.String())
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

	return IsNewerVersion(currentVersion, targetVersion)
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
