package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

func TestDownloadRauc(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "rauc")
	assert.NoError(t, err)
	assert.NotNil(t, release)

	assert.NotNil(t, release.Mirrors)
	assert.NotEmpty(t, release.Mirrors)

	assert.NotNil(t, release.Modules)
	assert.NotEmpty(t, release.Modules)

	assert.NotNil(t, release.Packages)
	assert.NotEmpty(t, release.Packages)

	assert.NotEmpty(t, release.ReleaseNotes)
	assert.NotEmpty(t, release.Version)

	tmpDir, err := os.MkdirTemp("", "casaos-installer-rauc-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)
	// fmt.Println(releaseFilePath)

	releaseFilePath, err = service.VerifyRAUC(*release)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)
}
func TestInstallRelease(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	assert.NotNil(t, release)

	assert.NotNil(t, release.Mirrors)
	assert.NotEmpty(t, release.Mirrors)

	assert.NotNil(t, release.Modules)
	assert.NotEmpty(t, release.Modules)

	assert.NotNil(t, release.Packages)
	assert.NotEmpty(t, release.Packages)

	assert.NotEmpty(t, release.ReleaseNotes)
	assert.NotEmpty(t, release.Version)

	tmpDir, err := os.MkdirTemp("", "casaos-installer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	fixtures.CacheRelease0441(config.ServerInfo.CachePath)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	assert.NoFileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos"))
	assert.NoFileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-gateway"))
	assert.NoFileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-app-management"))
	assert.NoFileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-local-storage"))

	err = service.InstallCasaOSPackages(*release, tmpSysRoot)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos"))
	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-gateway"))
	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-app-management"))
	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos-local-storage"))
}

func TestDownloadRelease(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-download-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)
	fmt.Println(releaseFilePath)
}

func TestPostReleaseInsall(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "casaos-installer-test-*")
	defer os.RemoveAll(tmpDir)

	assert.NoError(t, err)
	tmpSysRoot := filepath.Join(tmpDir, "sysroot")
	os.MkdirAll(tmpSysRoot, 0755)
	os.MkdirAll(filepath.Join(tmpSysRoot, "etc", "casaos"), 0755)

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	err = service.PostReleaseInstall(*release, tmpSysRoot)
	assert.NoError(t, err)

	// to check the target file is exist
	assert.FileExists(t, filepath.Join(tmpSysRoot, "etc", "casaos", "target-release.yaml"))
}

func TestIsUpgradable(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "casaos-update-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// to download release files
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
	assert.NoError(t, err)

	fixtures.SetLocalRelease(tmpSysRoot, "v0.4.5")

	result := service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, false)

	// mock /usr/bin/casaos
	// casaosVersion := "v0.4.3"
	fixtures.SetLocalRelease(tmpSysRoot, "v0.4.3")

	currentVersion, err := service.CurrentReleaseVersion(tmpSysRoot)
	assert.NoError(t, err)
	assert.Equal(t, currentVersion, "0.4.3")

	result = service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, true)

	// test case: the version can be update, but the package is not exist
	result = service.IsUpgradable(*release, tmpSysRoot)
	assert.Equal(t, result, false)

	fixtures.CacheRelease0441(config.ServerInfo.CachePath)

	_, err = service.VerifyRelease(*release)
	assert.NoError(t, err)

	// test case: the version can be update and the package is  exist
	result = service.IsUpgradable(*release, tmpSysRoot)
	assert.Equal(t, result, true)
}

func TestDeviceModelDiscover(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-device-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	// to download release files
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	resule := service.IsCasaOS(tmpSysRoot)
	assert.Equal(t, resule, true)
	resule = service.IsZimaOS(tmpSysRoot)
	assert.Equal(t, resule, false)

	fixtures.SetZimaOS(tmpSysRoot)
	assert.NoError(t, err)

	resule = service.IsCasaOS(tmpSysRoot)
	assert.Equal(t, resule, false)
	resule = service.IsZimaOS(tmpSysRoot)
	assert.Equal(t, resule, true)
}
