package service_test

import (
	"context"
	"fmt"
	"os"
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

func Test_Status_Case1_CRONJOB(t *testing.T) {
	// 测试说明: 老版本在就绪之后重新检测更新,并触发新的更新
	// 本地版本 老版本
	// 线上版本 新版本
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-1")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{
			InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
			DownloadStatusLock: sync.RWMutex{},
		},
		SysRoot:                          sysRoot,
		Have_other_get_release_flag_lock: sync.RWMutex{},
	}

	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	// main的过程
	statusService.Launch(sysRoot)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, "other", msg)

	statusService.PostMigration(sysRoot)

	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.UP_TO_DATE, msg)

	// 模拟cron
	go func() {
		ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)
		statusService.GetRelease(ctx, "latest")
		statusService.DownloadRelease(ctx, codegen.Release{}, false)
	}()

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.FetchUpdating, value.Status)
	assert.Equal(t, "fetching", msg)

	time.Sleep(2 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, types.DOWNLOADING, msg)

	time.Sleep(2 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.READY_TO_UPDATE, msg)
}

func Test_Status_Case2_HTTP_GET_Release(t *testing.T) {
	// 测试说明: 老版本在就绪之后重新检测更新,并触发新的更新
	// 本地版本 老版本
	// 线上版本 新版本
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)

	statusService.UpdateStatusWithMessage(service.DownloadEnd, types.READY_TO_UPDATE)
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, types.READY_TO_UPDATE, msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
	// 现在模仿HTTP请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// HTTP 请求的getRelease不会更新状态
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, string(types.READY_TO_UPDATE), msg)

	time.Sleep(2 * time.Second)
	// 但是应该会说需要更新
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, "http 触发的下载", msg)
}

func Test_Status_Case3_Install_Success(t *testing.T) {
	// 测试说明: 测试完整的安装流程
	// 本地版本 老版本
	// 线上版本 新版本

	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-3")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)
	// 模仿安装时的状态

	statusService.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.INSTALL)
	// 现在模仿install请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// 安装 请求的getRelease会把状态变成installing
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.FETCHING), msg)

	time.Sleep(3 * time.Second)
	go statusService.DownloadRelease(ctx, codegen.Release{}, false)

	time.Sleep(1 * time.Second)
	// 安装 请求的dowing会把状态变成installing
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DOWNLOADING), msg)

	go statusService.ExtractRelease("", codegen.Release{})

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DECOMPRESS), msg)

	go statusService.Install(codegen.Release{}, "")

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.INSTALLING), msg)
}

// 这个是测试cron触发的更新的下载的更新的
func Test_Status_Case2_Upgradable(t *testing.T) {
	logger.LogInitConsoleOnly()
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
		// 这个在github上的环境跳过测试，因为下载太快了，没有办法断言到downloading
	}

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)
	sysRoot := tmpDir
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := service.NewStatusService(&service.RAUCService{
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.OnlineRAUCExist,
		UrlHandler:         service.GitHubBranchTagReleaseUrl,
	}, sysRoot)

	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)

	statusService.UpdateStatusWithMessage(service.FetchUpdateEnd, "")

	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	_, err = statusService.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Downloading, value.Status)
	assert.Equal(t, "downloading", msg)

	time.Sleep(5 * time.Second)
	fmt.Println("断言")
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

}

func Test_Status_Case3_Download_Failed(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)
	sysRoot := tmpDir
	config.ServerInfo.CachePath = filepath.Join(tmpDir, "cache")

	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := service.NewStatusService(&service.RAUCService{
		InstallRAUCHandler: service.MockInstallRAUC,
		CheckSumHandler:    checksum.AlwaysFail,
		UrlHandler:         service.GitHubBranchTagReleaseUrl,
	}, sysRoot)

	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)

	statusService.UpdateStatusWithMessage(service.FetchUpdateEnd, "")

	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	go func() {
		_, err = statusService.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Microsecond)

	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.FetchUpdating, value.Status)
	assert.Equal(t, "fetching", msg)

	time.Sleep(10 * time.Second)

	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.DownloadError, value.Status)
	assert.Equal(t, "download fail", msg)
}

func Test_Status_Case4_Install_Fail(t *testing.T) {
	// 测试说明: 测试下载成功、安装时失败
	// 本地版本 老版本
	// 线上版本 新版本

	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-4")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysFailedInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)
	// 模仿安装时的状态

	statusService.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")
	value, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.INSTALL)
	// 现在模仿install请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// 安装 请求的getRelease会把状态变成installing
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.FETCHING), msg)

	time.Sleep(3 * time.Second)
	go statusService.DownloadRelease(ctx, codegen.Release{}, false)

	time.Sleep(1 * time.Second)
	// 安装 请求的dowing会把状态变成installing
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DOWNLOADING), msg)

	go statusService.ExtractRelease("", codegen.Release{})

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DECOMPRESS), msg)

	go statusService.Install(codegen.Release{}, "")

	time.Sleep(1 * time.Second)
	value, msg = statusService.GetStatus()
	assert.Equal(t, codegen.InstallError, value.Status)
	assert.Equal(t, "rauc is not compatible", msg)
}

func Test_Status_Get_Release_Currency(t *testing.T) {
	// 测试说明: 测试同时拿多次 release
	// 本地版本 老版本
	// 线上版本 新版本

	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-5")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.5")

	statusService := service.NewStatusService(&service.TestService{
		InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		DownloadStatusLock: sync.RWMutex{},
	}, sysRoot)
	statusService.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")

	service.Test_server_count_lock.Lock()
	service.ShouldUpgradeCount = 0
	service.Test_server_count_lock.Unlock()

	go func() {
		ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
		release, err := statusService.GetRelease(ctx, "latest")
		assert.NoError(t, err)
		assert.Equal(t, "v0.4.8", release.Version)
	}()

	time.Sleep(100 * time.Microsecond)

	go func() {
		ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
		release, err := statusService.GetRelease(ctx, "latest")
		assert.NoError(t, err)
		assert.Equal(t, "v0.4.8", release.Version)
	}()

	time.Sleep(100 * time.Microsecond)

	go func() {
		ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
		release, err := statusService.GetRelease(ctx, "latest")
		assert.NoError(t, err)
		assert.Equal(t, "v0.4.8", release.Version)
	}()

	time.Sleep(1 * time.Second)

	status, msg := statusService.GetStatus()
	assert.Equal(t, codegen.Idle, status.Status)
	assert.Equal(t, "ready-to-update", msg)

	time.Sleep(5 * time.Second)
	service.Test_server_count_lock.Lock()
	assert.Equal(t, 1, service.ShouldUpgradeCount)
	service.Test_server_count_lock.Unlock()

}
