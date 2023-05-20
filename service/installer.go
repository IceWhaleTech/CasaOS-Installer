package service

import (
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/labstack/echo/v4"
)

type InstallerService struct{}

func (i *InstallerService) GetLatest(ctx echo.Context) (codegen.Release, error) {
	panic("implement me")
}

func NewInstallerService() *InstallerService {
	return &InstallerService{}
}
