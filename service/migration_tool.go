package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/Masterminds/semver/v3"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type MigrationTool struct {
	Version semver.Version
	URL     string
}

func DownloadAllMigrationTools(ctx context.Context, release codegen.Release) (bool, error) {
	sourceVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		return false, err
	}

	migrationToolsMap, err := MigrationToolsMap(release)
	if err != nil {
		return false, err
	}

	downloaded := false

	for module, migrationTools := range migrationToolsMap {
		currentVersion, err := CurrentModuleVersion(module)
		if err != nil {
			logger.Info("failed to get the current version of module - skipping", zap.Error(err), zap.String("module", module))
			continue
		}

		if !sourceVersion.GreaterThan(currentVersion) {
			logger.Info("no need to migrate", zap.String("module", module), zap.String("sourceVersion", sourceVersion.String()), zap.String("currentVersion", currentVersion.String()))
			continue
		}

		for _, migration := range migrationTools {
			if migration.Version.LessThan(currentVersion) || migration.Version.GreaterThan(sourceVersion) {
				continue
			}

			if err := DownloadMigrationTool(ctx, release, module, migration); err != nil {
				return false, err
			}

			downloaded = true
		}
	}

	return downloaded, nil
}

func DownloadMigrationTool(ctx context.Context, release codegen.Release, module string, migration MigrationTool) error {
	template := NormalizeMigrationToolURL(migration.URL)

	outDir := filepath.Join(MigrationToolsDir(), module)

	for _, mirror := range release.Mirrors {
		url := strings.ReplaceAll(template, common.MirrorPlaceHolder, mirror)
		if _, err := internal.Download(ctx, outDir, url); err != nil {
			logger.Info("error while downloading migration tool - skipping", zap.Error(err), zap.String("url", migration.URL))
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to download migration tool %s", migration.URL)
}

// Normalize migraiton tool URL to a standard format which uses `${MIRROR}` as the mirror placeholder
func NormalizeMigrationToolURL(url string) string {
	url = NormalizeMigrationToolURLPass1(url)
	url = NormalizeMigrationToolURLPass2(url)

	url = strings.ReplaceAll(url, common.ArchPlaceHolder, lo.If(runtime.GOARCH == "arm", "arm-7").Else(runtime.GOARCH))
	return url
}

func NormalizeMigrationToolURLPass1(url string) string {
	// adapt to an old version of the migration list, where URL is just a version string
	//
	// e.g. CasaOS-Gateway/build/scripts/migration/service.d/gateway/migration.list
	//
	// LEGACY_WITHOUT_VERSION v0.3.6
	// v0.3.5 v0.3.6
	// v0.3.5.1 v0.3.6
	if _, err := semver.NewVersion(NormalizeVersion(url)); err != nil {
		return url
	}

	return fmt.Sprintf("%s/CasaOS/releases/download/%s/linux-%s-casaos-migration-tool-%s.tar.gz", common.MirrorPlaceHolder, url, common.ArchPlaceHolder, url)
}

func NormalizeMigrationToolURLPass2(url string) string {
	// adapt to an old version of the migration list, where URL assumes base path is ${DOWNLOAD_DOMAIN}IceWhaleTech
	//
	// e.g. CasaOS/build/scripts/migration/service.d/casaos/migration.list
	//
	// LEGACY_WITHOUT_VERSION ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz
	// v0.3.5 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz
	// v0.3.5.1 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz
	return strings.ReplaceAll(url, "${DOWNLOAD_DOMAIN}IceWhaleTech", common.MirrorPlaceHolder)
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

// verify migration tools for a release are already cached
func VerifyAllMigrationTools(release codegen.Release) bool {
	panic("implement me") // TODO
}
