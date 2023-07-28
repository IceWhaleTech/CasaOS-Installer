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

func TestDownloadMigrationTool(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	tmpDir, err := os.MkdirTemp("", "casaos-migration-test-*")
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

	tmpDir, err := os.MkdirTemp("", "casaos-migration-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "dev-test")
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
	fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.4.3")
	migrationPath, err := service.GetMigrationPath(module, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 0)

	fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.5")
	migrationPath, err = service.GetMigrationPath(module, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 1)
}

func TestMigrationPath(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-migration-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "dev-test")
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

	fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.0")

	migrationPath, err := service.GetMigrationPath(module, *release, migrationToolMap, tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, len(migrationPath), 2)
}

func TestDownloadAndInstallMigrateion(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-migration-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")
	os.Mkdir(tmpSysRoot, 0755)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	release, err := service.GetRelease(ctx, "dev-test")
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
	fixtures.SetCasaOSVersion(tmpSysRoot, module, "v0.3.5")
	migrationPath, err := service.GetMigrationPath(module, *release, migrationToolMap, tmpSysRoot)
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
