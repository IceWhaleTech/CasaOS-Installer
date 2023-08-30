package route

import (
	"context"
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var sysRoot = "/"

func (a *api) GetStatus(ctx echo.Context) error {
	status := service.GetStatus()
	return ctx.JSON(http.StatusOK, &codegen.StatusOK{
		Data:    &status,
		Message: nil,
	})
}

func (a *api) GetRelease(ctx echo.Context, params codegen.GetReleaseParams) error {

	tag := service.GetReleaseBranch(sysRoot)

	go service.PublishEventWrapper(context.Background(), common.EventTypeCheckUpdateBegin, nil)
	defer service.PublishEventWrapper(context.Background(), common.EventTypeCheckUpdateEnd, nil)
	go service.UpdateStatus(codegen.Status{
		Status: codegen.FetchUpdating,
	})
	defer service.UpdateStatus(codegen.Status{
		Status: codegen.FetchUpdated,
	})

	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	release, err := service.GetRelease(ctx.Request().Context(), tag)
	if err != nil {
		message := err.Error()
		service.PublishEventWrapper(context.Background(), common.EventTypeCheckUpdateError, map[string]string{
			common.PropertyTypeMessage.Name: err.Error(),
		})

		if err == service.ErrReleaseNotFound {
			return ctx.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
				Message: &message,
			})
		}

		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	upgradable := service.IsUpgradable(*release, "")

	return ctx.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data:       release,
		Upgradable: &upgradable,
	})
}

func (a *api) InstallRelease(ctx echo.Context, params codegen.InstallReleaseParams) error {
	go service.UpdateStatus(codegen.Status{
		Status: codegen.Installing,
	})
	defer service.UpdateStatus(codegen.Status{
		Status: codegen.Installed,
	})

	tag := service.GetReleaseBranch(sysRoot)

	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	release, err := service.InstallerService.GetRelease(ctx.Request().Context(), tag)
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

	if release == nil {
		message := "release not found"
		return ctx.JSON(http.StatusNotFound, &codegen.ResponseNotFound{
			Message: &message,
		})
	}

	go func() {
		go service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateBegin, nil)
		defer service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateEnd, nil)

		// backgroundCtx, cancel := context.WithCancel(context.Background())
		// defer cancel()
		sysRoot := "/"

		// if the err is not nil. It mean should to download
		contentCtx := context.Background()
		releasePath, err := service.InstallerService.DownloadRelease(contentCtx, *release, false)
		if err != nil {
			go service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateError, map[string]string{
				common.PropertyTypeMessage.Name: err.Error(),
			})

			logger.Error("error while download release: %s", zap.Error(err))
			return
		}

		// TODO disable migration when rauc install temporarily
		// // to download migration script
		// if _, err := service.DownloadAllMigrationTools(backgroundCtx, *release, sysRoot); err != nil {
		// 	go service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateError, map[string]string{
		// 		common.PropertyTypeMessage.Name: err.Error(),
		// 	})

		// 	logger.Error("error while download migration: %s", zap.Error(err))
		// 	// 回头这里重做一下，有rauc自己的migration
		// 	//return
		// }

		err = service.InstallerService.ExtractRelease(releasePath, *release)
		if err != nil {
			go service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateError, map[string]string{
				common.PropertyTypeMessage.Name: err.Error(),
			})

			logger.Error("error while extract release: %s", zap.Error(err))
			return
		}

		err = service.InstallerService.Install(*release, sysRoot)
		if err != nil {
			go service.PublishEventWrapper(context.Background(), common.EventTypeInstallUpdateError, map[string]string{
				common.PropertyTypeMessage.Name: err.Error(),
			})

			logger.Error("error while install system: %s", zap.Error(err))
			return
		}

	}()

	message := "release being installed asynchronously"
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
		Message: &message,
	})
}
