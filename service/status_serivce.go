package service

import (
	"context"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type StatusService struct {
	ImplementService InstallerServiceInterface
	SysRoot          string
}

func (r *StatusService) Install(release codegen.Release, sysRoot string) error {
	UpdateStatusWithMessage(InstallBegin, "installing")

	return r.ImplementService.Install(release, sysRoot)
}

func (r *StatusService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	release := &codegen.Release{}
	UpdateStatusWithMessage(FetchUpdateBegin, "触发更新")
	defer func() {
		if r.ShouldUpgrade(*release, r.SysRoot) {
			UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
			return
		} else {
			if r.IsUpgradable(*release, r.SysRoot) {
				UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
			} else {
				UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
			}
		}
		UpdateStatusWithMessage(FetchUpdateEnd, "触发更新")
	}()

	release, err := r.ImplementService.GetRelease(ctx, tag)
	return release, err
}

func (r *StatusService) MigrationInLaunch(sysRoot string) error {
	// 在这里会把状态更新为installing或者继续idle
	UpdateStatusWithMessage(InstallBegin, "migration")
	// defer UpdateStatusWithMessage(InstallEnd, "migration")
	return r.ImplementService.MigrationInLaunch(sysRoot)
}

func (r *StatusService) VerifyRelease(release codegen.Release) (string, error) {
	return r.ImplementService.VerifyRelease(release)
}

func (r *StatusService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	UpdateStatusWithMessage(DownloadBegin, "自动触发的下载")
	// TODO 这里想一下错误怎么处理?
	defer UpdateStatusWithMessage(DownloadEnd, "ready-to-update")
	return r.ImplementService.DownloadRelease(ctx, release, force)
}

func (r *StatusService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	UpdateStatusWithMessage(InstallBegin, "decompress")
	return r.ImplementService.ExtractRelease(packageFilepath, release)
}

func (r *StatusService) PostInstall(release codegen.Release, sysRoot string) error {
	UpdateStatusWithMessage(InstallBegin, "restarting")
	return r.ImplementService.PostInstall(release, sysRoot)
}

func (r *StatusService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return r.ImplementService.ShouldUpgrade(release, sysRoot)
}

func (r *StatusService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	return r.ImplementService.IsUpgradable(release, sysRootPath)
}

func (r *StatusService) GetMigrationInfo(ctx context.Context, release codegen.Release) error {

	return r.ImplementService.GetMigrationInfo(ctx, release)
}

func (r *StatusService) DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	return r.ImplementService.DownloadAllMigrationTools(ctx, release)
}

func (r *StatusService) PostMigration(sysRoot string) error {
	UpdateStatusWithMessage(InstallEnd, "up-to-date")
	return r.ImplementService.PostMigration(sysRoot)
}

func (r *StatusService) Cronjob(sysRoot string) error {
	return nil
}
