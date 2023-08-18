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
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestNormalizationVersion(t *testing.T) {
	assert.Equal(t, "v0.4.2-1", service.NormalizeVersion("v0.4.2-1"))
	assert.Equal(t, "v0.4.2-1", service.NormalizeVersion("v0.4.2.1"))
	assert.Equal(t, "v0.4.2", service.NormalizeVersion("v0.4.2"))

	assert.Equal(t, "v0.4.2-1", service.NormalizationVersion("v0.4.2-1"))
	assert.Equal(t, "v0.4.2-1", service.NormalizationVersion("v0.4.2.1"))
	assert.Equal(t, "v0.4.2", service.NormalizationVersion("v0.4.2"))

}

func TestIsUpgradableInSpecifyVersion(t *testing.T) {

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

	// mock /usr/bin/casaos
	// casaosVersion := "v0.4.3"
	fixtures.SetLocalRelease(tmpSysRoot, "v0.4.2-1")

	currentVersion, err := service.CurrentReleaseVersion(tmpSysRoot)
	assert.NoError(t, err)
	version, err := semver.NewVersion("0.4.2-1")
	assert.NoError(t, err)
	assert.Equal(t, true, currentVersion.Equal(version))

	result := service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, true)

	fixtures.SetLocalRelease(tmpSysRoot, "v0.4.2.1")

	currentVersion, err = service.CurrentReleaseVersion(tmpSysRoot)
	assert.NoError(t, err)

	version, err = semver.NewVersion("0.4.2-1")
	assert.NoError(t, err)
	assert.Equal(t, true, currentVersion.Equal(version))

	result = service.ShouldUpgrade(*release, tmpSysRoot)
	assert.Equal(t, result, true)
}
