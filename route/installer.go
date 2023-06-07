package route

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/labstack/echo/v4"
)

func (a *api) GetRelease(ctx echo.Context, params codegen.GetReleaseParams) error {
	tag := "main"
	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	release, err := service.GetRelease(tag)
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
		Data: release,
	})
}

func (a *api) InstallRelease(ctx echo.Context, params codegen.InstallReleaseParams) error {
	tag := "main"
	if params.Version != nil && *params.Version != "latest" {
		tag = *params.Version
	}

	release, err := service.GetRelease(tag)
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

	if err := service.InstallRelease(ctx, *release); err != nil {
		message := err.Error()
		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	message := "release being installed asynchronously"
	return ctx.JSON(http.StatusOK, &codegen.ResponseOK{
		Message: &message,
	})
}
