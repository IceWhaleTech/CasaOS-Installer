package service

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

type MigrationTool struct {
	version semver.Version
	url     string
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

		if !sourceVersion.GreaterThan(&currentVersion) {
			logger.Info("no need to migrate", zap.String("module", module), zap.String("sourceVersion", sourceVersion.String()), zap.String("currentVersion", currentVersion.String()))
			continue
		}

		for _, migration := range migrationTools {
			if migration.version.LessThan(&currentVersion) || migration.version.GreaterThan(sourceVersion) {
				continue
			}

			if err := DownloadMigrationTool(ctx, releaseDir, migration); err != nil {
				return err
			}

			// TODO
		}
	}

	panic("implement me")
}

func DownloadMigrationTool(ctx context.Context, releaseDir string, migration MigrationTool) error {
	panic("implement me")
}

func CurrentVersion(module string) (semver.Version, error) {
	panic("implement me")
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
				version: *version,
				url:     parts[1],
			})
		}

	}

	return migrationToolsMap, nil
}
