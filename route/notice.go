package route

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

func (a *api) GetNoticeStatus(c echo.Context) error {
	ctx := c.Request().Context()
	tag := service.GetReleaseBranch(config.SysRoot)
	release, err := service.InstallerService.GetRelease(ctx, tag)
	if release == nil {
		return c.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: lo.ToPtr("release is fetching"),
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &codegen.ResponseInternalServerError{
			Message: lo.ToPtr(err.Error()),
		})
	}
	_, packageStatus := service.InstallerService.GetStatus()

	switch packageStatus {
	case types.READY_TO_UPDATE:
		if release.Important != nil && *release.Important {
			return c.JSON(http.StatusOK, &codegen.NoticeStatusOK{
				Data: &codegen.NoticeStatus{
					Status: codegen.ImportantUpdate,
				},
			})
		}
		return c.JSON(http.StatusOK, &codegen.NoticeStatusOK{
			Data: &codegen.NoticeStatus{
				Status: codegen.NormalUpdate,
			},
		})
	default:
		return c.JSON(http.StatusOK, &codegen.NoticeStatusOK{
			Data: &codegen.NoticeStatus{
				Status: codegen.NoUpdate,
			},
		})
	}
}
