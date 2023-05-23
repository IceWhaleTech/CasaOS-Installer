package service

import (
	"fmt"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/labstack/echo/v4"
)

type InstallerService struct{}

func (i *InstallerService) GetRelease(ctx echo.Context, tag string) (*codegen.Release, error) {
	releaseURL := fmt.Sprintf("%s/%s/casaos-release", config.ServerInfo.ReleaseBaseURL, tag)

	return internal.GetReleaseFrom(releaseURL)
}

func (i *InstallerService) InstallRelease(ctx echo.Context, release codegen.Release) error {
	return internal.InstallRelease(release)
}

func NewInstallerService() *InstallerService {
	return &InstallerService{}
}
