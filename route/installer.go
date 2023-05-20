package route

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/labstack/echo/v4"
)

func (a *api) GetLatest(ctx echo.Context) error {
	release, err := service.Installer.GetLatest(ctx)
	if err != nil {
		message := err.Error()
		return ctx.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: &message,
		})
	}

	return ctx.JSON(http.StatusOK, &codegen.ReleaseOK{
		Data: &release,
	})
}
