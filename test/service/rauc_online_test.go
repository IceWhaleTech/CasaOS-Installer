package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
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
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OnlineTarExist,
	}

	release, err := installerServer.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)
	assert.Equal(t, "v0.4.4-1", release.Version)
	fmt.Println(release)
	// 这个是一个假文件，只有2.6mb
	releasePath, err := installerServer.DownloadRelease(ctx, *release, false)
	parentDir := filepath.Dir(releasePath)

	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.tar"))
	// run shell in golang

	releasePath, err = installerServer.VerifyRelease(*release)
	assert.NoError(t, err)

	err = installerServer.ExtractRelease(releasePath, *release)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.raucb"))

	// ensure release file exists
	fmt.Println("release path:", releasePath)
	assert.FileExists(t, releasePath)

	// ensure rauc file exists
	// get parent dir of releaseDir
	assert.FileExists(t, filepath.Join(parentDir, "casaos_ova-0.4.4-1.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}
