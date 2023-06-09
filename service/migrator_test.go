package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
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

	version, err := semver.NewVersion("v0.4.4")
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.DownloadMigrationTool(ctx, tmpDir, "casaos", service.MigrationTool{
		Version: *version,
		URL:     "asdf",
	})

	assert.Error(t, err)
}
