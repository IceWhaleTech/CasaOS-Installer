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
	config.SysRoot = tmpDir
	config.ServerInfo.Mirrors = []string{
		"https://raw.githubusercontent.com/IceWhaleTech/get/main/",
	}

	installerServer := &service.RAUCService{
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OnlineTarExist,
		UrlHandler:         service.GitHubBranchTagReleaseUrl,
	}

	release, err := installerServer.GetRelease(ctx, "unit-test-rauc-online-v2-0.5.0")
	assert.NoError(t, err)
	assert.Equal(t, "v0.5.0", release.Version)
	fmt.Println(release)

	releasePath, err := installerServer.DownloadRelease(ctx, *release, false)
	parentDir := filepath.Dir(releasePath)

	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "zimaos_zimacube-0.5.0.raucb"))
	// run shell in golang

	releasePath, err = installerServer.VerifyRelease(*release)
	assert.NoError(t, err)

	err = installerServer.ExtractRelease(releasePath, *release)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(parentDir, "zimaos_zimacube-0.5.0.raucb"))

	// ensure release file exists
	fmt.Println("release path:", releasePath)
	assert.FileExists(t, releasePath)

	// ensure rauc file exists
	// get parent dir of releaseDir
	assert.FileExists(t, filepath.Join(parentDir, "zimaos_zimacube-0.5.0.raucb"))

	err = installerServer.Install(*release, tmpDir)
	assert.NoError(t, err)
}
