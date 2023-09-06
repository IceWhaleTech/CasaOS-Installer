package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/stretchr/testify/assert"
)

func Test_Status_Case1_Launch_have_Update(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-1")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{},
		SysRoot:          sysRoot,
	}

	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	// main的过程
	statusService.MigrationInLaunch(sysRoot)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, "migration", msg)

	statusService.PostMigration(sysRoot)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "up-to-date", msg)

	// 模拟cron
	go func() {
		statusService.GetRelease(context.TODO(), "latest")
		statusService.DownloadRelease(context.TODO(), codegen.Release{}, false)
	}()

	time.Sleep(1 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.FetchUpdating, value.Status)
	assert.Equal(t, "触发更新", msg)

	time.Sleep(2 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, "下载中", msg)

	time.Sleep(2 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)
}

func Test_Status_Case2_Upgradable(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)
	sysRoot := tmpDir
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := &service.StatusService{
		ImplementService: &service.RAUCService{
			InstallRAUCHandler: service.InstallRAUCTest,
		},
		SysRoot: sysRoot,
	}

	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	release, err := statusService.GetRelease(context.TODO(), "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "out-of-date", msg)

	_, err = statusService.DownloadRelease(context.TODO(), *release, false)
	assert.NoError(t, err)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

}
