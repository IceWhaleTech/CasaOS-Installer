package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
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
	assert.Equal(t, "间隔触发更新", msg)

	time.Sleep(3 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, "自动触发的下载", msg)

	time.Sleep(2 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)
}

func Test_Status_Case2_Install(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)
	sysRoot := tmpDir

	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	_ = &service.StatusService{
		ImplementService: &service.TestService{},
		SysRoot:          sysRoot,
	}

}
