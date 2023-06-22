package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

func MigrationToolsDir() string {
	return filepath.Join(config.ServerInfo.CachePath, "migration-tools")
}

func ReleaseDir(release codegen.Release) (string, error) {
	if release.Version == "" {
		return "", fmt.Errorf("release version is empty")
	}

	return filepath.Join(config.ServerInfo.CachePath, "releases", release.Version), nil
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

func CurrentReleaseVersion() (*semver.Version, error) {
	// TODO: look for the release info first before looking for the binary version (legacy)
	return CurrentModuleVersion("casaos")
}

func CurrentModuleVersion(module string) (*semver.Version, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	for _, executablePath := range []string{
		"/usr/bin/" + module,
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
