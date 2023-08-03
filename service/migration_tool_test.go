package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	// TODO NOTE! there is a bug, the url after v0.3.5-1 didn't download the migration tool
	appManagementMigrationList = `v0.3.5 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS-AppManagement/releases/download/v0.4.0/linux-${ARCH}-casaos-app-management-migration-tool-v0.4.0.tar.gz
v0.3.6 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS-AppManagement/releases/download/v0.4.1-alpha1/linux-${ARCH}-casaos-app-management-migration-tool-v0.4.1-alpha1.tar.gz
v0.3.7 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS-AppManagement/releases/download/v0.4.2-1/linux-${ARCH}-casaos-app-management-migration-tool-v0.4.2-1.tar.gz
v0.3.8 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS-AppManagement/releases/download/v0.4.3/linux-${ARCH}-casaos-app-management-migration-tool-v0.4.3.tar.gz
v0.3.9 ${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS-AppManagement/releases/download/v0.4.3/linux-${ARCH}-casaos-app-management-migration-tool-v0.4.3.tar.gz`
)

// NOTE! the test will cost very long time(1 min)(decided by network speed). So we should timeout it longer than another.
//
//	/usr/local/go/bin/go test -timeout 290s -run ^TestDownloadAllMigrationTools$ github.com/IceWhaleTech/CasaOS-Installer/service
func TestDownloadAllMigrationTools(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	tmpDir, err := os.MkdirTemp("", "casaos-download-all-migration-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = tmpDir

	logger.LogInitConsoleOnly()

	targetVersionRelease, err := service.GetRelease(context.Background(), "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")
	os.Mkdir(tmpSysRoot, 0755)

	fixtures.SetLocalRelease(tmpSysRoot, "v0.3.5")

	// to construct a fake migration map
	releaseDir, err := service.ReleaseDir(*targetVersionRelease)
	assert.NoError(t, err)
	migrationListDir := filepath.Join(releaseDir, "build/scripts/migration/service.d")
	for _, module := range targetVersionRelease.Modules {
		migrationListFile := filepath.Join(migrationListDir, module.Short, common.MigrationListFileName)
		err = os.MkdirAll(filepath.Dir(migrationListFile), 0755)
		fmt.Println(filepath.Dir(migrationListFile))
		assert.NoError(t, err)
		// to write a fake migration list file
		err = os.WriteFile(migrationListFile, []byte(appManagementMigrationList), 0644)
		assert.NoError(t, err)
	}

	_, err = service.DownloadAllMigrationTools(context.Background(), *targetVersionRelease, tmpSysRoot)
	assert.NoError(t, err)

	// to check if all migration tools are downloaded
	// it should be 4 migration tools

	AppManagementOutDir := filepath.Join(service.MigrationToolsDir(), "casaos-app-management")
	fmt.Println(AppManagementOutDir)

	assert.DirExists(t, AppManagementOutDir)
	// find out the files in the directory
	files, err := os.ReadDir(AppManagementOutDir)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(files))
}

func TestDownloadMigrationTool(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	tmpDir, err := os.MkdirTemp("", "casaos-download-migration-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = tmpDir

	logger.LogInitConsoleOnly()

	version, err := semver.NewVersion(service.NormalizeVersion("v0.4.3"))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var release codegen.Release

	err = yaml.Unmarshal([]byte(common.SampleReleaseYAML), &release)
	assert.NoError(t, err)

	DOWNLOAD_DOMAIN := "https://github.com/"
	module := "CasaOS-AppManagement"
	moduleShort := "casaos-app-management"
	arch := "amd64"
	versionTag := "v0.4.3"
	migrationDownloadURL := fmt.Sprintf(
		"%sIceWhaleTech/%s/releases/download/%s/linux-%s-%s-migration-tool-%s.tar.gz",
		DOWNLOAD_DOMAIN,
		module,
		versionTag,
		arch,
		moduleShort,
		versionTag,
	)
	path, err := service.DownloadMigrationTool(ctx, release, module, service.MigrationTool{
		Version: *version,
		URL:     migrationDownloadURL,
	}, false)
	fmt.Println(path)
	assert.NoError(t, err)
}

func TestMigrationToolsMap(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-migration-map-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)

	err = service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	// extract very module package that the name is like linux*.tar.gz
	err = service.ExtractReleasePackages(releaseFilePath+"/linux*", *release)
	assert.NoError(t, err)

	migrationToolMap, err := service.MigrationToolsMap(*release)
	assert.NoError(t, err)
	fmt.Println(migrationToolMap)

	module := "casaos-local-storage"
	fixtures.SetLocalRelease(tmpSysRoot, "v0.4.3")
	// fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.4.3")
	migrationPath, err := service.GetMigrationPath(codegen.Module{
		Short: module,
		Name:  module,
	}, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 0)

	fixtures.SetLocalRelease(tmpSysRoot, "v0.3.5")
	// fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.5")
	migrationPath, err = service.GetMigrationPath(codegen.Module{
		Short: module,
		Name:  module,
	}, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 1)
}

func TestMigrationPath(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-migration-path-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	module := "casaos-local-storage"

	migrationToolMap := map[string][]service.MigrationTool{}
	migrationToolMap[module] = []service.MigrationTool{
		{
			Version: *semver.MustParse("0.3.0"),
			URL:     "download 0.3.5 script",
		},
		{
			Version: *semver.MustParse("0.3.5"),
			URL:     "download 0.4.0 script",
		},
		{
			Version: *semver.MustParse("0.4.5"),
			URL:     "download 0.5.0 script",
		},
	}

	fixtures.SetLocalRelease(tmpSysRoot, "v0.3.0")
	// fixtures.SetCasaOSVersion(tmpSysRoot, "casaos", "v0.3.0")
	// fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.0")

	migrationPath, err := service.GetMigrationPath(codegen.Module{
		Short: module,
		Name:  module,
	}, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 2)
}

func TestDownloadAndInstallMigrateion(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-execute-migration-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")
	os.Mkdir(tmpSysRoot, 0755)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)

	err = service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	// extract very module package that the name is like linux*.tar.gz
	err = service.ExtractReleasePackages(releaseFilePath+"/linux*", *release)
	assert.NoError(t, err)

	migrationToolMap, err := service.MigrationToolsMap(*release)
	assert.NoError(t, err)

	module := "casaos-local-storage"
	fixtures.SetLocalRelease(tmpSysRoot, "v0.3.5")
	fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.5")
	migrationPath, err := service.GetMigrationPath(codegen.Module{
		Short: module,
		Name:  module,
	}, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 1)

	for _, migration := range migrationPath {
		migrationPath, err := service.DownloadMigrationTool(ctx, *release, module, migration, false)
		assert.NoError(t, err)
		err = service.ExecuteMigrationTool(module, migrationPath, tmpSysRoot)
		// because MigrationTool require root permission, so it will return exit status 1
		assert.Equal(t, err.Error(), "exit status 1")
	}
}
