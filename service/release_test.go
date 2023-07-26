package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

var casaos043VersionScript = "#! /usr/bin/python3\nprint(\"v0.4.3\")"
var casaos045VersionScript = "#! /usr/bin/python3\nprint(\"v0.4.5\")"

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
	defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)

	err = service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	// extract very module package that the name is like linux*.tar.gz
	err = service.ExtractReleasePackages(releaseFilePath+"/linux*", *release)
	assert.NoError(t, err)

	// downloaded, err := service.DownloadAllMigrationTools(ctx, *release)
	// assert.NoError(t, err)
	// assert.True(t, downloaded)

	fmt.Println("下载到", releaseFilePath)
	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	err = service.InstallRelease(ctx, *release, tmpSysRoot)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpSysRoot, "usr", "bin", "casaos"))
}

// the test require root permission
func TestIsUpgradable(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := os.Stat("/tmp/usr/bin"); os.IsNotExist(err) {
		// to create folder
		err := os.Mkdir("/tmp/usr", 0755)
		assert.NoError(t, err)
		err = os.Mkdir("/tmp/usr/bin", 0755)
		assert.NoError(t, err)
	}

	release, err := service.GetRelease(ctx, "dev-test")
	assert.NoError(t, err)
	// TODO - to for more easy to test. such config casaos position
	casaosPath := filepath.Join("/tmp", "usr", "bin", "casaos")

	// mock /usr/bin/casaos
	// casaosVersion := "v0.4.5"
	err = os.WriteFile(casaosPath, []byte(casaos045VersionScript), 0755)
	assert.NoError(t, err)
	defer os.Remove(casaosPath)

	result := service.ShouldUpgrade(*release)
	assert.Equal(t, result, false)

	// mock /usr/bin/casaos
	// casaosVersion := "v0.4.3"
	err = os.WriteFile(casaosPath, []byte(casaos043VersionScript), 0755)
	assert.NoError(t, err)

	result = service.ShouldUpgrade(*release)
	assert.Equal(t, result, true)

	//to delete /var/cache/casaos. ensure the casaos package is not exist
	_ = os.RemoveAll("/var/cache/casaos")

	result = service.IsUpgradable(*release)
	assert.Equal(t, result, false)

	// to download release files
	tmpDir, err := os.MkdirTemp("", "casaos-installer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	if _, err := os.Stat("/var/cache/casaos"); os.IsNotExist(err) {
		err = os.Mkdir("/var/cache/casaos", 0755)
		defer os.RemoveAll("/var/cache/casaos")
		assert.NoError(t, err)
	}

	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)

	service.ExtractReleasePackages(releaseFilePath, *release)
	assert.NoError(t, err)

	_, err = service.VerifyRelease(*release)
	assert.NoError(t, err)

	result = service.IsUpgradable(*release)
	assert.Equal(t, result, true)
}
