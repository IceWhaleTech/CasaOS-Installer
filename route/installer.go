package route

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Common/utils"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var sysRoot = "/"

func (a *api) GetStatus(ctx echo.Context) error {
	status, packageStatus := service.GetStatus()
	return ctx.JSON(http.StatusOK, &codegen.StatusOK{
		Data:    &status,
		Message: utils.Ptr(packageStatus),
	})
}

func (a *api) GetRelease(ctx echo.Context, params codegen.GetReleaseParams) error {
	status, _ := service.GetStatus()
	if status.Status == codegen.Downloading {
		message := "downloading"
		return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
			Message: &message,
		})
	}
	if status.Status == codegen.Installing {
		message := "downloading"
		return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
			Message: &message,
		})
	}

	// TODO 考虑一下这个packageStatus的问题
	// go service.UpdateStatusWithMessage(service.FetchUpdateBegin, "主动触发的获取信息")
	tag := service.GetReleaseBranch(sysRoot)

	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	release, err := service.InstallerService.GetRelease(context.WithValue(ctx.Request().Context(), types.Trigger, types.HTTP_REQUEST), tag)
	if err != nil {
		message := err.Error()
		if err == service.ErrReleaseNotFound {
			return ctx.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
				Message: &message,
			})
		}
		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	upgradable := service.InstallerService.IsUpgradable(*release, "")

	if service.ShouldUpgrade(*release, sysRoot) {
		if upgradable {
			service.UpdateStatusWithMessage(service.FetchUpdateEnd, "ready-to-update")
		} else {
			service.UpdateStatusWithMessage(service.FetchUpdateEnd, "out-of-date")
			service.InstallerService.DownloadRelease(context.Background(), *release, false)
		}
	} else {
		service.UpdateStatusWithMessage(service.FetchUpdateEnd, "up-to-date")
	}

	return ctx.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data:       release,
		Upgradable: &upgradable,
	})
}

func (a *api) InstallRelease(ctx echo.Context, params codegen.InstallReleaseParams) error {
	status, _ := service.GetStatus()
	if status.Status == codegen.Downloading {
		message := "downloading"
		return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
			Message: &message,
		})
	}

	if status.Status == codegen.Installing {
		message := "installing"
		return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
			Message: &message,
		})
	}

	// go service.UpdateStatusWithMessage(service.InstallBegin, "主动触发的安装更新1级")

	tag := service.GetReleaseBranch(sysRoot)

	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	// go service.UpdateStatusWithMessage(service.InstallBegin, "getRelease中")

	release, err := service.InstallerService.GetRelease(context.WithValue(ctx.Request().Context(), types.INSTALL, types.INSTALL), tag)
	if err != nil {
		message := err.Error()

		if err == service.ErrReleaseNotFound {
			return ctx.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
				Message: &message,
			})
		}

		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	go service.UpdateStatusWithMessage(service.InstallBegin, "回复完成")

	if release == nil {
		message := "release not found"
		return ctx.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
			Message: &message,
		})
	}

	go func() {
		// backgroundCtx, cancel := context.WithCancel(context.Background())
		// defer cancel()
		sysRoot := "/"

		// if the err is not nil. It mean should to download
		contentCtx := context.Background()

		releasePath, err := service.InstallerService.DownloadRelease(contentCtx, *release, false)
		if err != nil {
			logger.Error("error while downloading release: %s", zap.Error(err))
			service.UpdateStatusWithMessage(service.InstallError, fmt.Sprintf("安装时下载失败:%s", err.Error()))
			return
		}

		fmt.Println("解压目录:", releasePath)
		err = service.InstallerService.ExtractRelease(releasePath, *release)
		if err != nil {
			logger.Error("error while extract release: %s", zap.Error(err))
			service.UpdateStatusWithMessage(service.InstallError, fmt.Sprintf("安装时解压失败:%s", err.Error()))
			return
		}

		err = service.InstallerService.Install(*release, sysRoot)
		if err != nil {
			logger.Error("error while install system: %s", zap.Error(err))
			service.UpdateStatusWithMessage(service.InstallError, fmt.Sprintf("安装失败:%s", err.Error()))
			return
		}

		err = service.InstallerService.PostInstall(*release, sysRoot)
		if err != nil {
			logger.Error("error while post install system: %s", zap.Error(err))
			service.UpdateStatusWithMessage(service.InstallError, fmt.Sprintf("后安装失败:%s", err.Error()))
			return
		}
	}()

	message := "release being installed asynchronously"
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
		Message: &message,
	})
}

func (a *api) ResetStatus(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{})
}
