package route

import (
	"context"
	"net/http"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (a *api) GetBackground(ctx echo.Context, param codegen.GetBackgroundParams) error {
	// no cache
	ctx.Response().Header().Set("Cache-Control", "no-cache")
	return ctx.File(internal.BackgroundPath(*param.Version))
}

func (a *api) GetStatus(ctx echo.Context) error {
	status, packageStatus := service.InstallerService.GetStatus()
	return ctx.JSON(http.StatusOK, &codegen.StatusOK{
		Data:    &status,
		Message: utils.Ptr(packageStatus),
	})
}

func (a *api) GetRelease(c echo.Context, params codegen.GetReleaseParams) error {
	tag := service.GetReleaseBranch(config.SysRoot)
	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	ctx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
	release, err := service.InstallerService.GetRelease(ctx, tag)

	if err != nil {
		message := err.Error()
		if err == service.ErrReleaseNotFound {
			return c.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
				Message: &message,
			})
		}
		return c.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	// TODO refactor this
	// the code might be remove
	if release.Background == nil {
		logger.Error("release.Background is nil")
	} else {
		go internal.DownloadReleaseBackground(*release.Background, release.Version)
	}

	release.Background = utils.Ptr("/v2/installer/background?version=" + release.Version)

	return c.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data:       release,
		Upgradable: nil,
	})
}

func (a *api) InstallRelease(ctx echo.Context, params codegen.InstallReleaseParams) error {
	status, _ := service.InstallerService.GetStatus()

	service.InstallerService.UpdateStatusWithMessage(service.InstallBegin, types.FETCHING)

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

		ctx := context.WithValue(context.Background(), types.Trigger, types.INSTALL)
		release, err := service.InstallerService.GetRelease(ctx, tag)
		if err != nil {
			message := err.Error()
			service.InstallerService.UpdateStatusWithMessage(service.InstallError, message)
			return
		}

		if release == nil {
			service.InstallerService.UpdateStatusWithMessage(service.InstallError, "release is nil")
			return
		}

		// if the err is not nil. It mean should to download

		releasePath, err := service.InstallerService.DownloadRelease(ctx, *release, false)
		if err != nil {
			service.InstallerService.UpdateStatusWithMessage(service.InstallError, err.Error())
			logger.Error("error while downloading release: %s")
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

// 这里是重置状态，但是没有用到，因为改成在上面getRelease也能重置状态。但是后续也可能会用到
func (a *api) ResetStatus(ctx echo.Context) error {
	installCtx := context.WithValue(ctx.Request().Context(), types.Trigger, types.HTTP_REQUEST)
	service.InstallerService.GetRelease(installCtx, "latest")
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{})
}

func (a *api) GetInstall(ctx echo.Context) error {
	tag := service.GetReleaseBranch(config.SysRoot)

	installCtx := context.WithValue(context.Background(), types.Trigger, types.HTTP_REQUEST)
	release, err := service.InstallerService.GetRelease(installCtx, tag)

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

	path, err := service.InstallerService.InstallInfo(*release, config.SysRoot)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: utils.Ptr(err.Error()),
		})
	}
	return ctx.JSON(http.StatusOK, &codegen.InstallInfoOk{
		Path: &path,
	})
}
