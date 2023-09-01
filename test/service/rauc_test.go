package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

// TODO 把这个修一下
// func TestVerifyRAUC(t *testing.T) {
// 	logger.LogInitConsoleOnly()

// 	tmpDir, err := os.MkdirTemp("", "casaos-verify-rauc-*")
// 	defer os.RemoveAll(tmpDir)
// 	assert.NoError(t, err)
// 	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
// 	assert.NoError(t, err)

// 	_, err = service.VerifyRAUC(*release)
// 	assert.ErrorContains(t, err, "not found")

// 	// TODO to download rauc

// 	// TODO to verify rauc again
// }

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

	releasePath, err = installerServer.VerifyRelease(*release)
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
