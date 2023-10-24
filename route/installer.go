package route

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (a *api) GetStatus(ctx echo.Context) error {
	status, packageStatus := service.GetStatus()
	return ctx.JSON(http.StatusOK, &codegen.StatusOK{
		Data:    &status,
		Message: utils.Ptr(packageStatus),
	})
}

func (a *api) GetRelease(ctx echo.Context, params codegen.GetReleaseParams) error {
	tag := service.GetReleaseBranch(config.SysRoot)
	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	http_trigger_context := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
	release, err := service.InstallerService.GetRelease(http_trigger_context, tag)

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

	return ctx.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data:       release,
		Upgradable: nil,
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

	go func() {
		tag := service.GetReleaseBranch(config.SysRoot)

		if params.Version != nil && *params.Version != "latest" {
			tag = *params.Version
		}

		release, err := service.InstallerService.GetRelease(context.WithValue(ctx.Request().Context(), types.INSTALL, types.INSTALL), tag)
		if err != nil {
			message := err.Error()
			fmt.Println("安装失败", message)
		}

		// if the err is not nil. It mean should to download
		contentCtx := context.Background()

		releasePath, err := service.InstallerService.DownloadRelease(contentCtx, *release, false)
		if err != nil {
			logger.Error("error while downloading release: %s", zap.Error(err))
			return
		}
		time.Sleep(3 * time.Second)

		err = service.InstallerService.ExtractRelease(releasePath, *release)
		if err != nil {
			logger.Error("error while extract release: %s", zap.Error(err))
			return
		}
		time.Sleep(3 * time.Second)

		err = service.InstallerService.Install(*release, config.SysRoot)
		if err != nil {
			logger.Error("error while install system: %s", zap.Error(err))
			return
		}

		err = service.InstallerService.PostInstall(*release, config.SysRoot)
		if err != nil {
			logger.Error("error while post install system: %s", zap.Error(err))
			return
		}
	}()

	message := "release being installed asynchronously"
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
		Message: &message,
	})
}

func (a *api) ResetStatus(ctx echo.Context) error {
	http_trigger_context := context.WithValue(ctx.Request().Context(), types.Trigger, types.HTTP_REQUEST)
	service.InstallerService.GetRelease(http_trigger_context, "latest")
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{})
}
