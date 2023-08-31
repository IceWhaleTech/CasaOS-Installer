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

func TestRAUCServer(t *testing.T) {

	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-rauc-download-extract-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	assert.NoError(t, err)
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	installerServer := &service.RAUCService{
		InstallRAUCHandler: service.InstallRAUCTest,
	}

	release, err := installerServer.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

	// 这个是一个假文件，只有2.6mb
	releasePath, err := installerServer.DownloadRelease(ctx, *release, false)
	parentDir := filepath.Dir(releasePath)

	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.tar.gz"))

	releasePath, err = service.VerifyRelease(*release)
	assert.NoError(t, err)

	err = installerServer.ExtractRelease(releasePath, *release)
	assert.NoError(t, err)

	// ensure release file exists
	assert.FileExists(t, filepath.Join(releasePath))

	// ensure rauc file exists
	// get parent dir of releaseDir
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}

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

	os.MkdirAll(filepath.Join(tmpDir, service.RAUCOfflinePath), 0755)
	fixtures.SetOfflineRAUC(tmpDir, service.RAUCOfflinePath, service.RAUCOfflineRAUCFile)

	release, err := installerServer.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

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
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}
