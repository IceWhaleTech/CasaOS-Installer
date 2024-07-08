package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

func TestRAUCOfflineServer(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-rauc-offline-extract-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	assert.NoError(t, err)
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	installerServer := &service.RAUCOfflineService{
		SysRoot:            tmpDir,
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OfflineTarExistV2,
		GetRAUCInfo:        service.MockRAUCInfo,
	}

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")
	config.SysRoot = tmpDir

	// 构建假文件放到目录

	config.RAUC_OFFLINE_RAUC_FILENAME = "rauc.raucb"

	os.MkdirAll(filepath.Join(tmpDir, config.RAUC_OFFLINE_PATH), 0o755)
	fixtures.SetOfflineRAUC(tmpDir, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
	fixtures.SetOfflineRAUCMock_0504(tmpDir)

	release, err := installerServer.GetRelease(ctx, "any thing")
	assert.NoError(t, err)

	assert.Equal(t, "v0.5.0.4", release.Version)
	assert.Equal(t, "# private test\n", release.ReleaseNotes)

	// 这个是一个假文件，只有2.6mb
	releasePath, err := installerServer.DownloadRelease(ctx, *release, false)
	parentDir := filepath.Dir(releasePath)
	assert.NoError(t, err)

	_, err = installerServer.VerifyRelease(*release)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(parentDir, "rauc.raucb"))

	err = installerServer.ExtractRelease(releasePath, *release)
	assert.NoError(t, err)

	// ensure release file exists
	assert.FileExists(t, filepath.Join(releasePath))

	// ensure rauc file exists
	// get parent dir of releaseDir
	assert.FileExists(t, filepath.Join(parentDir, "rauc.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}

func TestRAUCOfflineServerLoadReleaseFromCache(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-rauc-offline-extract-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	assert.NoError(t, err)
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	installerServer := &service.RAUCOfflineService{
		SysRoot:            tmpDir,
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OfflineTarExistV2,
		GetRAUCInfo:        service.MockRAUCInfo,
	}

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")
	config.SysRoot = tmpDir

	// 构建假文件放到目录

	config.RAUC_OFFLINE_RAUC_FILENAME = "rauc.raucb"

	os.MkdirAll(filepath.Join(tmpDir, config.RAUC_OFFLINE_PATH), 0o755)
	fixtures.SetOfflineRAUC(tmpDir, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
	fixtures.SetOfflineRAUCMock_0504(tmpDir)
	fixtures.SetOfflineRAUCRelease_050(tmpDir)
	assert.FileExists(t, filepath.Join(tmpDir, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME))

	release, err := installerServer.GetRelease(ctx, "any thing")
	assert.NoError(t, err)

	assert.Equal(t, "v0.5.0", release.Version)
	assert.Equal(t, "# private test\n", release.ReleaseNotes)
}

func TestRAUCOfflineServerGetReleaseFail(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-rauc-offline-extract-test-*")
	assert.NoError(t, err)
	// defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	assert.NoError(t, err)
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	installerServer := &service.RAUCOfflineService{
		SysRoot:            tmpDir,
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OfflineTarExistV2,
		GetRAUCInfo:        service.MockRAUCInfo,
	}

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")
	config.SysRoot = tmpDir

	// 构建假文件放到目录

	config.RAUC_OFFLINE_RAUC_FILENAME = "rauc.raucb"

	os.MkdirAll(filepath.Join(tmpDir, config.RAUC_OFFLINE_PATH), 0o755)
	fixtures.SetOfflineRAUC(tmpDir, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
	fixtures.SetOfflineRAUCMock_049(tmpDir)

	_, err = installerServer.GetRelease(ctx, "any thing")
	assert.ErrorContains(t, err, "illegal base64 data")
}
