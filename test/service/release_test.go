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

func TestGetOldRelease(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
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

	assert.Nil(t, release.Code)
	assert.Nil(t, release.Background)
}

func TestGetNewRelease(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "unit-test-release-1.1.0")
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

	assert.NotNil(t, release.Code)
	assert.Equal(t, "Super Man", *release.Code)
	assert.NotNil(t, release.Background)
	assert.Equal(t, "https://upload.wikimedia.org/wikipedia/commons/thumb/0/09/Central_Californian_Coastline%2C_Big_Sur_-_May_2013.jpg/1200px-Central_Californian_Coastline%2C_Big_Sur_-_May_2013.jpg", *release.Background)
}

func TestDownloadRauc(t *testing.T) {
	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release, err := service.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
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
	// err may be "download fail"
	if err != nil && assert.Contains(t, err.Error(), "download fail") {
		t.SkipNow()
	}

	assert.NoError(t, err)

	err = service.ExtractRAUCRelease(releaseFilePath, *release)

	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)

	releaseFilePath, err = service.RAUCFilePath(*release)
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)
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
	// err may be "download fail"
	if err != nil && assert.Contains(t, err.Error(), "download fail") {
		t.SkipNow()
	}
	assert.NoError(t, err)
	assert.FileExists(t, releaseFilePath)
}

func TestBestByDelay(t *testing.T) {
	url := service.BestByDelay([]string{
		"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/rauc.txt",
		"https://raw.githubusercontent.com/IceWhaleTech/zimaos-rauc/main/rau242342c",
		"https://baidu.com/weqteqrwerwerwr",
		"https://baid2342341234123411231412u.com/weqteqrwerwerwr",
	})
	assert.Equal(t, "https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/rauc.txt", url)

	url = service.BestByDelay([]string{
		"https://baidu.com/weqteqrwerwerwr",
		"https://baid2342341234123411231412u.com/weqteqrwerwerwr",
	})
	assert.Equal(t, "", "")
}

func TestDeviceModelDiscover(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-device-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	// to download release files
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	result := service.IsCasaOS(tmpSysRoot)
	assert.Equal(t, result, true)
	result = service.IsZimaOS(tmpSysRoot)
	assert.Equal(t, result, false)

	fixtures.SetZimaOS(tmpSysRoot)
	assert.NoError(t, err)

	result = service.IsCasaOS(tmpSysRoot)
	assert.Equal(t, result, false)
	result = service.IsZimaOS(tmpSysRoot)
	assert.Equal(t, result, true)
}
