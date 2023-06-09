package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

type MigrationTool struct {
	Version semver.Version
	URL     string
}

func DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	sourceVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		return err
	}

	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	migrationToolsMap, err := MigrationToolsMap(release)
	if err != nil {
		return err
	}

	for module, migrationTools := range migrationToolsMap {
		currentVersion, err := CurrentVersion(module)
		if err != nil {
			return err
		}

		if !sourceVersion.GreaterThan(currentVersion) {
			logger.Info("no need to migrate", zap.String("module", module), zap.String("sourceVersion", sourceVersion.String()), zap.String("currentVersion", currentVersion.String()))
			continue
		}

		for _, migration := range migrationTools {
			if migration.Version.LessThan(currentVersion) || migration.Version.GreaterThan(sourceVersion) {
				continue
			}

			if err := DownloadMigrationTool(ctx, releaseDir, module, migration); err != nil {
				return err
			}

			// TODO
		}
	}

	panic("implement me")
}

func DownloadMigrationTool(ctx context.Context, releaseDir string, module string, migration MigrationTool) error {
	migrationToolURL, err := NormalizeMigrationToolURL(migration.URL)
	if err != nil {
		return err
	}

	migrationToolsDir := filepath.Join(releaseDir, "migration", module)

	// TODO: fill the URL template, e.g. ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz

	return internal.DownloadAndExtractPackage(ctx, migrationToolsDir, migrationToolURL)
}

func NormalizeMigrationToolURL(url string) (string, error) {
	if !strings.HasSuffix(url, ".tar.gz") { // some old migration list has no full URL template, but just a version
		url = NormalizeMigrationToolURLPass1(url)
	}

	panic("implement me")
}

func NormalizeMigrationToolURLPass1(url string) string {
	panic("implement me")
}

func CurrentVersion(module string) (*semver.Version, error) {
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

func MigrationToolsMap(release codegen.Release) (map[string][]MigrationTool, error) {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return nil, err
	}

	// TODO: auto detect migrationListDir in future, instead of hardcoding it
	migrationListDir := filepath.Join(releaseDir, "build/scripts/migration/service.d")

	migrationToolsMap := map[string][]MigrationTool{}

	for _, module := range release.Modules {
		migrationListFile := filepath.Join(migrationListDir, module.Short, common.MigrationListFileName)

		file, err := os.Open(migrationListFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		migrationToolsMap[module.Name] = []MigrationTool{}

		scanner := bufio.NewScanner(file)

		// iterate over lines
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "#") {
				logger.Info("skipping comment line", zap.String("line", line))
				continue
			}

			parts := strings.Fields(line)

			if len(parts) != 2 {
				logger.Info("invalid migration list line", zap.String("line", line))
				continue
			}

			parts[0] = NormalizeVersion(parts[0])

			version, err := semver.NewVersion(parts[0])
			if err != nil {
				return nil, err
			}

			migrationToolsMap[module.Name] = append(migrationToolsMap[module.Name], MigrationTool{
				Version: *version,
				URL:     parts[1],
			})
		}

	}

	return migrationToolsMap, nil
}
