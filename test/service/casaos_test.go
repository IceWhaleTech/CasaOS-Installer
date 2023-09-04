package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

func TestUninstallScript(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-uninstall-script-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	value := service.VerifyUninstallScript(tmpSysRoot)
	assert.Equal(t, false, value)

	ctx := context.Background()
	service.DownloadUninstallScript(ctx, tmpSysRoot)

	value = service.VerifyUninstallScript(tmpSysRoot)
	assert.Equal(t, true, value)
}
