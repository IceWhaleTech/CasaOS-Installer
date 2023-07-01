package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDownloadMigrationTool(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	tmpDir, err := os.MkdirTemp("", "casaos-installer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config.ServerInfo.CachePath = tmpDir

	logger.LogInitConsoleOnly()

	version, err := semver.NewVersion(service.NormalizeVersion("v0.3.5.1"))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var release codegen.Release

	err = yaml.Unmarshal([]byte(common.SampleReleaseYAML), &release)
	assert.NoError(t, err)

	_, err = service.DownloadMigrationTool(ctx, release, "casaos", service.MigrationTool{
		Version: *version,
		URL:     "${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz",
	}, false)

	assert.NoError(t, err)
}
