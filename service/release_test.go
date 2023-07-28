package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

func TestInstallRelease(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "dev-test")
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
	// defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)

	err = service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	// extract very module package that the name is like linux*.tar.gz
	err = service.ExtractReleasePackages(releaseFilePath+"/linux*", *release)
	assert.NoError(t, err)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	err = service.InstallRelease(ctx, *release, tmpSysRoot)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos"))
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

	release, err := service.GetRelease(ctx, "dev-test")
	assert.NoError(t, err)

	fixtures.SetCasaOSVersion(tmpSysRoot, "casaos", "v0.4.5")

	result := service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, false)

	// mock /usr/bin/casaos
	// casaosVersion := "v0.4.3"
	fixtures.SetCasaOSVersion(tmpSysRoot, "casaos", "v0.4.3")

	result = service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, true)

	// test case: the version can be update, but the package is not exist
	result = service.IsUpgradable(*release, tmpSysRoot)
	assert.Equal(t, result, false)

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)

	service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	_, err = service.VerifyRelease(*release)
	assert.NoError(t, err)

	// test case: the version can be update and the package is  exist
	result = service.IsUpgradable(*release, tmpSysRoot)
	assert.Equal(t, result, true)
}
