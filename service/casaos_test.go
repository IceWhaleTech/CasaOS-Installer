package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/stretchr/testify/assert"
)

func TestUninstallScript(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-uninstall-script-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpSysRoot := filepath.Join(tmpDir, "sysroot")

	value := VerifyUninstallScript(tmpSysRoot)
	assert.Equal(t, false, value)

	ctx := context.Background()
	DownloadUninstallScript(ctx, tmpSysRoot)

	value = VerifyUninstallScript(tmpSysRoot)
	assert.Equal(t, true, value)
}
