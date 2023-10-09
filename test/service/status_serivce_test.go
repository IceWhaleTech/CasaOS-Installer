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
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/stretchr/testify/assert"
)

func Test_Status_Case1_CRONJOB(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-1")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.4")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{
			InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		},
		SysRoot: sysRoot,
	}

	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	// main的过程
	statusService.Launch(sysRoot)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, "other", msg)

	statusService.PostMigration(sysRoot)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "up-to-date", msg)

	// 模拟cron
	go func() {
		ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)
		statusService.GetRelease(ctx, "latest")
		statusService.DownloadRelease(ctx, codegen.Release{}, false)
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

func Test_Status_Case2_HTTP_GET_Release(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-2")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{
			InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		},
		SysRoot: sysRoot,
	}

	service.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")
	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
	// 现在模仿HTTP请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// HTTP 请求的getRelease不会更新状态
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, string(types.READY_TO_UPDATE), msg)

	time.Sleep(3 * time.Second)
	// 但是应该会说需要更新
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, string(types.OUT_OF_DATE), msg)

}

func Test_Status_Case3_Install_Success(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-3")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{
			InstallRAUCHandler: service.AlwaysSuccessInstallHandler,
		},
		SysRoot: sysRoot,
	}
	// 模仿安装时的状态

	service.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")
	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.INSTALL)
	// 现在模仿install请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// 安装 请求的getRelease会把状态变成installing
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.FETCHING), msg)

	time.Sleep(3 * time.Second)
	go statusService.DownloadRelease(ctx, codegen.Release{}, false)

	time.Sleep(1 * time.Second)
	// 安装 请求的dowing会把状态变成installing
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DOWNLOADING), msg)

	go statusService.ExtractRelease("", codegen.Release{})

	time.Sleep(1 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DECOMPRESS), msg)

	go statusService.Install(codegen.Release{}, "")

	time.Sleep(1 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.INSTALLING), msg)
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
			InstallRAUCHandler: service.MockInstallRAUC,
			CheckSumHandler:    checksum.OnlineTarExist,
			UrlHandler:         service.GitHubBranchTagReleaseUrl,
		},
		SysRoot: sysRoot,
	}

	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)

	service.UpdateStatusWithMessage(service.FetchUpdateEnd, "")

	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	release, err := statusService.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, string(types.OUT_OF_DATE), msg)

	_, err = statusService.DownloadRelease(ctx, *release, false)
	assert.NoError(t, err)

	value, msg = service.GetStatus()
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

	statusService := &service.StatusService{
		ImplementService: &service.RAUCService{
			InstallRAUCHandler: service.MockInstallRAUC,
			CheckSumHandler:    checksum.AlwaysFail,
			UrlHandler:         service.GitHubBranchTagReleaseUrl,
		},
		SysRoot: sysRoot,
	}

	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	ctx := context.WithValue(context.Background(), types.Trigger, types.CRON_JOB)

	service.UpdateStatusWithMessage(service.FetchUpdateEnd, "")

	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "", msg)

	release, err := statusService.GetRelease(ctx, "unit-test-rauc-0.4.4-1")
	assert.NoError(t, err)

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, string(types.OUT_OF_DATE), msg)

	_, err = statusService.DownloadRelease(ctx, *release, false)
	assert.ErrorContains(t, err, "download fail")

	value, msg = service.GetStatus()
	assert.Equal(t, codegen.DownloadError, value.Status)
	assert.Equal(t, "download fail", msg)
}

func Test_Status_Case4_Install_Fail(t *testing.T) {
	logger.LogInitConsoleOnly()

	tmpDir, err := os.MkdirTemp("", "casaos-status-test-case-4")
	assert.NoError(t, err)

	sysRoot := tmpDir
	fixtures.SetLocalRelease(sysRoot, "v0.4.3")

	statusService := &service.StatusService{
		ImplementService: &service.TestService{
			InstallRAUCHandler: service.AlwaysFailedInstallHandler,
		},
		SysRoot: sysRoot,
	}
	// 模仿安装时的状态

	service.UpdateStatusWithMessage(service.DownloadEnd, "ready-to-update")
	value, msg := service.GetStatus()
	assert.Equal(t, codegen.Idle, value.Status)
	assert.Equal(t, "ready-to-update", msg)

	ctx := context.WithValue(context.Background(), types.Trigger, types.INSTALL)
	// 现在模仿install请求拿更新
	go statusService.GetRelease(ctx, "latest")

	time.Sleep(1 * time.Second)
	// 安装 请求的getRelease会把状态变成installing
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.FETCHING), msg)

	time.Sleep(3 * time.Second)
	go statusService.DownloadRelease(ctx, codegen.Release{}, false)

	time.Sleep(1 * time.Second)
	// 安装 请求的dowing会把状态变成installing
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DOWNLOADING), msg)

	go statusService.ExtractRelease("", codegen.Release{})

	time.Sleep(1 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.Installing, value.Status)
	assert.Equal(t, string(types.DECOMPRESS), msg)

	go statusService.Install(codegen.Release{}, "")

	time.Sleep(1 * time.Second)
	value, msg = service.GetStatus()
	assert.Equal(t, codegen.InstallError, value.Status)
	assert.Equal(t, "rauc is not compatible", msg)
}

// TODO 补一个测试，就是解压失败
