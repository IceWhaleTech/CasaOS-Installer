package service

import (
	"context"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type InstallerServiceInterface interface {
	GetRelease(ctx context.Context, tag string) (*codegen.Release, error)
	VerifyRelease(release codegen.Release) (string, error)
	DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error)
	ExtractRelease(packageFilepath string, release codegen.Release) error
	GetMigrationInfo(ctx context.Context, release codegen.Release) error
	DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error
	Install(release codegen.Release, sysRoot string) error
	PostInstall(release codegen.Release, sysRoot string) error

	MigrationInLaunch(sysRoot string) error
	PostMigration(sysRoot string) error

	ShouldUpgrade(release codegen.Release, sysRoot string) bool
	IsUpgradable(release codegen.Release, sysRootPath string) bool // 检测预下载的包好了没有
}
