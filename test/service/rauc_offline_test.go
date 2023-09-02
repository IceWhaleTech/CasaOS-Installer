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
		InstallRAUCHandler: service.InstallRAUCTest,
	}

	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	// 构建假文件放到目录

	os.MkdirAll(filepath.Join(tmpDir, service.RAUC_OFFLINE_PATH), 0755)
	fixtures.SetOfflineRAUC(tmpDir, service.RAUC_OFFLINE_PATH, service.RAUC_OFFLINE_RAUC_FILENAME)

	release, err := installerServer.GetRelease(ctx, "any thing")
	assert.NoError(t, err)

	assert.Equal(t, "v0.4.8", release.Version)
	assert.Equal(t, "rauc offline update test package", release.ReleaseNotes)

	assert.FileExists(t, filepath.Join(tmpDir, service.RAUC_OFFLINE_PATH, service.RAUC_OFFLINE_RAUC_FILENAME))
	assert.FileExists(t, filepath.Join(tmpDir, service.OFFLINE_RAUC_TEMP_PATH, service.RAUC_OFFLINE_RELEASE_FILENAME))

	// 这个是一个假文件，只有2.6mb
	releasePath, err := installerServer.DownloadRelease(ctx, *release, false)
	parentDir := filepath.Dir(releasePath)
	fmt.Println("下载目录:", releasePath)

	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "rauc.tar.gz"))

	err = installerServer.ExtractRelease(releasePath, *release)
	assert.NoError(t, err)

	// ensure release file exists
	assert.FileExists(t, filepath.Join(releasePath))

	// ensure rauc file exists
	// get parent dir of releaseDir
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.8.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}
