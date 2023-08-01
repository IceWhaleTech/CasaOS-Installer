package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
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

var (
	currentReleaseLocalPath = "/etc/casaos/release.yaml"
	targetReleaseLocalPath  = "/etc/casaos/release.yaml"
)

func StartMigration(sysRoot string) error {
	// check if migration is needed
	// current version
	currentVersion, err := CurrentReleaseVersion(sysRoot)
	if err != nil {
		return err
	}

	// target version
	targetVersion, err := semver.NewVersion(NormalizeVersion("v0.4.5"))
	if err != nil {
		return err
	}

	currentRelease, err := internal.GetReleaseFromLocal(currentReleaseLocalPath)
	if err != nil {
		return err
	}
	targetRelease, err := internal.GetReleaseFromLocal(targetReleaseLocalPath)
	if err != nil {
		return err
	}

	if currentVersion.GreaterThan(targetVersion) || currentVersion.Equal(targetVersion) {
		// no need to migrate
		return nil
	}

	// TODO: how to handle migration if the module of target version is more thatn the current version

	// start migration
	// 1. download migration tools
	// the migration tools should be downloaded when install release
	// So there is no need to download the migration tools again
	DownloadAllMigrationTools(context.Background(), *targetRelease, sysRoot)
	// for all modules of current release

	// 2. run migration tools
	migrationToolMap, err := MigrationToolsMap(*targetRelease)
	if err != nil {
		return err
	}

	for _, module := range currentRelease.Modules {
		migrationPath, err := GetMigrationPath(module, *targetRelease, migrationToolMap, sysRoot)
		if err != nil {
			return err
		}

		for _, migration := range migrationPath {
			// the migration tool should be downloaded when install release
			migrationPath, err := DownloadMigrationTool(context.Background(), *targetRelease, module.Short, migration, false)
			if err != nil {
				return err
			}
			err = ExecuteMigrationTool(module.Short, migrationPath, sysRoot)
			if err != nil {
				return err
			}
		}
	}

	// post migration
	return nil
}

func PostMigration(sysRoot string) error {
	// TODO: post migration. e.g. move target-relase to relase yaml
	return nil
}

func DownloadAllMigrationTools(ctx context.Context, release codegen.Release, sysrootPath string) (bool, error) {
	targetVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		return false, err
	}

	migrationToolsMap, err := MigrationToolsMap(release)
	if err != nil {
		return false, err
	}

	downloaded := false

	for module, migrationTools := range migrationToolsMap {
		currentVersion, err := CurrentReleaseVersion(sysrootPath)
		if err != nil {
			logger.Info("failed to get the current version of module - skipping", zap.Error(err), zap.String("module", module))
			continue
		}

		if !targetVersion.GreaterThan(currentVersion) {
			logger.Info("no need to migrate", zap.String("module", module), zap.String("targetVersion", targetVersion.String()), zap.String("currentVersion", currentVersion.String()))
			continue
		}

		for _, migration := range migrationTools {
			if migration.Version.LessThan(currentVersion) || migration.Version.GreaterThan(targetVersion) {
				continue
			}

			if path, err := DownloadMigrationTool(ctx, release, module, migration, false); err != nil {
				fmt.Println("下载完成", path)
				return false, err
			}

			downloaded = true
		}
	}

	return downloaded, nil
}

func DownloadMigrationTool(ctx context.Context, release codegen.Release, module string, migration MigrationTool, force bool) (string, error) {
	if !force {
		if migrationToolFilePath, err := VerifyMigrationTool(module, filepath.Base(migration.URL)); err != nil {
			logger.Info("error while verifying migration tool - continue to download", zap.Error(err))
		} else {
			return migrationToolFilePath, nil
		}
	}

	template := NormalizeMigrationToolURL(migration.URL)

	outDir := filepath.Join(MigrationToolsDir(), module)

	for _, mirror := range release.Mirrors {
		migrationToolURL := strings.ReplaceAll(template, common.MirrorPlaceHolder, mirror)
		migrationToolFilePath, err := internal.Download(ctx, outDir, migrationToolURL)
		if err != nil {
			logger.Info("error while downloading migration tool - skipping", zap.Error(err), zap.String("url", migrationToolURL))
			continue
		}

		// TODO: download checksums.txt and save the checksum for the migration tool to the same directory

		return migrationToolFilePath, nil
	}

	return "", fmt.Errorf("failed to download migration tool %s", migration.URL)
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

func GetMigrationPath(module codegen.Module, release codegen.Release, migrationToolMap map[string][]MigrationTool, sysRoot string) ([]MigrationTool, error) {
	sourceVersion, err := semver.NewVersion(NormalizeVersion(release.Version))
	if err != nil {
		return []MigrationTool{}, err
	}
	currentVersion, err := CurrentReleaseVersion(sysRoot)
	fmt.Println("currentVersion:::::", currentVersion)
	if err != nil {
		return []MigrationTool{}, err
	}

	PathArray := []MigrationTool{}

	modulePath := migrationToolMap[module.Short]
	for _, migration := range modulePath {
		if migration.Version.LessThan(sourceVersion) && (migration.Version.GreaterThan(currentVersion) || migration.Version.Equal(currentVersion)) {
			PathArray = append(PathArray, migration)
			// return migration.URL
		}
	}
	return RemoveDuplication(PathArray), nil
}

func ExecuteMigrationTool(module string, migrationFilePath string, sysRoot string) error {
	// to extract the migration tool
	err := internal.Extract(migrationFilePath, MigrationToolsDir())
	if err != nil {
		return err
	}

	// err = systemctl.StopService(module)
	// if err != nil {
	// 	return err
	// }

	// to execute the migration tool
	migrationToolPath := filepath.Join(MigrationToolsDir(), "build", "sysroot", "usr", "bin", module+"-migration-tool")
	// to chmod file permission
	err = os.Chmod(migrationToolPath, 0755)
	if err != nil {
		return err
	}
	// to execute the migration tool
	// force to execute the migration tool. otherwise, it require to stop the service
	cmd := exec.Command(migrationToolPath, "-f")
	fmt.Println(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// verify migration tools for a release are already cached
func VerifyAllMigrationTools(targetRelease codegen.Release, sysRoot string) bool {
	// get all migration tool
	currentRelease, err := internal.GetReleaseFromLocal(targetReleaseLocalPath)
	if err != nil {
		fmt.Println("获取release.yaml失败", currentRelease)
		return false
	}

	migrationToolMap, err := MigrationToolsMap(targetRelease)
	if err != nil {
		fmt.Println("获取migration map失败")

		return false
	}

	for _, module := range currentRelease.Modules {
		migrationPath, err := GetMigrationPath(module, targetRelease, migrationToolMap, sysRoot)
		if err != nil {
			fmt.Println("获取migration path失败", err)
			return false
		}

		for _, migration := range migrationPath {
			// the migration tool should be downloaded when install release
			_, err := VerifyMigrationTool(module.Short, NormalizeMigrationToolURL(filepath.Base(migration.URL)))
			if err != nil {
				return false
			}
		}
	}

	return true
}

func VerifyMigrationTool(module string, fileName string) (string, error) {
	migrationToolDir := filepath.Join(MigrationToolsDir(), module)

	packageFilePath := filepath.Join(migrationToolDir, fileName)

	// to check if the migration tool is already downloaded, we need to check if the file exists and its size
	if _, err := os.Stat(packageFilePath); err != nil {
		// TODO - verify the hash
		return "", fmt.Errorf("migration tool %s not found", packageFilePath)
	}

	return packageFilePath, nil
}
