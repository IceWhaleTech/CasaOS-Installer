package service_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common/fixtures"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/stretchr/testify/assert"
)

// 测试项目说明
// 这里是状态测试，功能代码是mock。只测试状态是否正确
func Test_Status_Case1_NotUpdate(t *testing.T) {
	// 测试说明: 成功下载测试

	// 本地版本 新版本
	// 线上版本 老版本
	logger.LogInitConsoleOnly()

	sysRoot := t.TempDir()
	ctx := context.Background()
	fixtures.SetLocalRelease(sysRoot, "v99.9.9")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)

	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	// 模拟cron
	go func() {
		statusService.Cronjob(ctx, sysRoot)
	}()

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.FetchUpdating, value.Status)
	assert.Equal(t, types.FETCHING, msg)

	fixtures.WaitFecthReleaseCompeleted(statusService)
	fixtures.WaitDownloadCompeleted(statusService)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.UP_TO_DATE, msg)
}

func Test_Status_Case1_Download_Success(t *testing.T) {
	// 测试说明: 成功下载测试
	// 成之后，状态应该是ready-to-update

	// 本地版本 老版本
	// 线上版本 新版本
	logger.LogInitConsoleOnly()

	sysRoot := t.TempDir()
	ctx := context.Background()
	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)

	// 模拟cron
	go func() {
		statusService.Cronjob(ctx, sysRoot)
	}()

	time.Sleep(1 * time.Second)
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.FetchUpdating, value.Status)
	assert.Equal(t, types.FETCHING, msg)

	fixtures.WaitFecthReleaseCompeleted(statusService)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, types.DOWNLOADING, msg)

	fixtures.WaitDownloadCompeleted(statusService)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.READY_TO_UPDATE, msg)
}

func Test_Status_Case2_Download_Failed(t *testing.T) {
	// 测试说明: 测试下载失败,下载之后无法通过checksum
	// 本地版本 老版本
	// 线上版本 新版本
	logger.LogInitConsoleOnly()

	sysRoot := t.TempDir()
	config.ServerInfo.CachePath = filepath.Join(sysRoot, "cache")
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")
	fixtures.SetZimaOS(sysRoot)

	statusService := service.NewStatusService(&service.RAUCService{
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.AlwaysFail,
		URLHandler:         service.GitHubBranchTagReleaseUrl,
	}, sysRoot)

	go func() {
		statusService.Cronjob(context.Background(), sysRoot)
	}()

	fixtures.WaitFecthReleaseCompeleted(statusService)
	fixtures.WaitDownloadCompeleted(statusService)

	value, msg := statusService.GetStatus()

	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "download fail", msg)
}

func Test_Status_Case3_Install_Success(t *testing.T) {
	// 测试说明: 测试在下载成功后，安装成功
	// 本地版本 老版本
	// 线上版本 新版本

	logger.LogInitConsoleOnly()

	sysRoot := t.TempDir()
	ctx := context.Background()
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)
	// 模仿安装时的状态

	go func() {
		statusService.Cronjob(ctx, sysRoot)
	}()

	fixtures.WaitFecthReleaseCompeleted(statusService)
	fixtures.WaitDownloadCompeleted(statusService)
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.READY_TO_UPDATE, msg)

	release, err := statusService.GetRelease(ctx, "latest")
	assert.NoError(t, err)
	go statusService.Install(*release, sysRoot)

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.INSTALLING), msg)
}

func Test_Status_Case4_Install_Fail(t *testing.T) {
	// 测试说明: 测试下载成功、安装时失败
	// 本地版本 老版本
	// 线上版本 新版本

	logger.LogInitConsoleOnly()

	sysRoot := t.TempDir()
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysFailedInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)
	// 模仿安装时的状态

	// TODO 重构这里用统一的就绪的fixtures
	statusService.UpdateStatusWithMessage(service.DownloadEnd, types.READY_TO_UPDATE)
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.READY_TO_UPDATE, msg)

	go func() {
		statusService.Install(codegen.Release{}, sysRoot)
	}()

	time.Sleep(5 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.InstallError, value.Status)
	assert.Equal(t, "rauc is not compatible", msg)
}
