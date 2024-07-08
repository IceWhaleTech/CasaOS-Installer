package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

func MigrationToolsDir() string {
	return filepath.Join(config.ServerInfo.CachePath, "migration-tools")
}

func NormalizeVersion(version string) string {
	if version == common.LegacyWithoutVersion {
		return "v0.0.0-legacy-without-version"
	}

	version = "v" + strings.TrimLeft(version, "Vv")

	if _, err := semver.NewVersion(version); err == nil {
		return version
	}

	versionNumbers := strings.SplitN(version, ".", 3)
	if len(versionNumbers) < 3 {
		return version
	}

	versionNumbers[2] = strings.Replace(versionNumbers[2], ".", "-", 1)

	return strings.Join([]string{versionNumbers[0], versionNumbers[1], versionNumbers[2]}, ".")
}

func CurrentReleaseVersion(sysRootPath string) (*semver.Version, error) {
	// get version from local release file, if not exist, get version from local casaos
	currentRelease, err := internal.GetReleaseFromLocal(filepath.Join(sysRootPath, CurrentReleaseLocalPath))
	if err != nil {
		return CurrentModuleVersion("zimaos", sysRootPath)
	}
	return semver.NewVersion(NormalizationVersion(currentRelease.Version))
}

func NormalizationVersion(version string) string {
	// v0.4.2.1 => v0.4.2-1
	// v0.4.2.1 => v0.4.2-1
	// v0.4.2 => v0.4.2

	versionNumbers := strings.SplitN(version, ".", 3)
	if len(versionNumbers) < 3 {
		return version
	}
	return strings.Join([]string{versionNumbers[0], versionNumbers[1], strings.Replace(versionNumbers[2], ".", "-", 1)}, ".")
}

func CurrentModuleVersion(module string, sysrootPath string) (*semver.Version, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	for _, executablePath := range []string{
		// in test environment, sysrootPath is in tmp like `/tmp/casaos-installer-test-*`
		// in production environment, the sysrootPath should be ``, abd the executable is /usr/bin
		sysrootPath + "/usr/bin/" + module,
		module,
	} {
		cmd := exec.Command(executablePath, "-v")
		if cmd == nil {
			continue
		}

		cmd.Stdout = writer

		logger.Info(cmd.String())

		if err := cmd.Run(); err != nil {
			logger.Info("failed to run command", zap.String("cmd", cmd.String()), zap.Error(err))
			continue
		}

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			logger.Info(line, zap.String("module", module))

			version, err := semver.NewVersion(NormalizeVersion(line))
			if err != nil {
				continue
			}

			return version, nil
		}
	}

	return nil, fmt.Errorf("failed to get current version of %s", module)
}

// func RemoveDuplication(arr []MigrationTool) []MigrationTool {
// 	length := len(arr)
// 	if length == 0 {
// 		return arr
// 	}

// 	j := 0
// 	for i := 1; i < length; i++ {
// 		if arr[i].URL != arr[j].URL {
// 			j++
// 			if j < i {
// 				swap(arr, i, j)
// 			}
// 		}
// 	}

// 	return arr[:j+1]
// }

// func swap(arr []MigrationTool, a, b int) {
// 	arr[a], arr[b] = arr[b], arr[a]
// }
