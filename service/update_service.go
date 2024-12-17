package service

import (
	"context"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

// add more info like update count or anything.
// TODO, it is helpful for debug
type UpdateServerStats struct {
	Name string
}

type UpdaterServiceInterface interface {
	GetRelease(ctx context.Context, tag string, useCache bool) (*codegen.Release, error)
	VerifyRelease(release codegen.Release) (string, error)
	DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error)
	ExtractRelease(packageFilepath string, release codegen.Release) error

	Install(release codegen.Release, sysRoot string) error
	PostInstall(release codegen.Release, sysRoot string) error

	Launch(sysRoot string) error
	PostMigration(sysRoot string) error

	ShouldUpgrade(release codegen.Release, sysRoot string) bool
	IsUpgradable(release codegen.Release, sysRootPath string) bool // 检测预下载的包好了没有

	InstallInfo(release codegen.Release, sysRoot string) (string, error)

	Stats() UpdateServerStats
}
