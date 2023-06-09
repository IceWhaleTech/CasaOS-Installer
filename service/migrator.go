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

type VersionURLMap struct {
	version semver.Version
	url     string
}

func DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	// TODO: auto detect migrationListDir in future, instead of hardcoding it
	migrationListDir := filepath.Join(releaseDir, "build/scripts/migration/service.d")

	migrationListMap := map[string][]VersionURLMap{}

	for _, module := range release.Modules {
		migrationListFile := filepath.Join(migrationListDir, module.Short, common.MigrationListFileName)

		file, err := os.Open(migrationListFile)
		if err != nil {
			return err
		}
		defer file.Close()

		migrationListMap[module.Name] = []VersionURLMap{}

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

			if strings.ToUpper(parts[0]) == common.LegacyWithoutVersion {
				parts[0] = "v0.0.0-legacy-without-version"
			}

			version, err := semver.NewVersion(strings.TrimLeft(parts[0], "vV"))
			if err != nil {
				return err
			}

			migrationListMap[module.Name] = append(migrationListMap[module.Name], VersionURLMap{
				version: *version,
				url:     parts[1],
			})
		}

	}

	panic("implement me")
}
