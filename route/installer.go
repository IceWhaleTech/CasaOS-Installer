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

	release, err := service.Installer.GetRelease(ctx, tag)
	if err != nil {
		message := err.Error()
		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	return ctx.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data: release,
	})
}

func (a *api) InstallRelease(ctx echo.Context, params codegen.InstallReleaseParams) error {
	panic("not implemented")
}
