package service

import (
	"context"
	"fmt"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
)

type StatusService struct {
	ImplementService UpdaterServiceInterface
	SysRoot          string
}

type InstallProgressStatus string

func (r *StatusService) Install(release codegen.Release, sysRoot string) error {
	UpdateStatusWithMessage(InstallBegin, types.INSTALLING)
	err := r.ImplementService.Install(release, sysRoot)
	defer func() {
		if err != nil {
			UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	release := &codegen.Release{}

	err := error(nil)
	// 因为更新完进入主页又要拿一次release
	if ctx.Value(types.Trigger) == types.HTTP_CHECK {
		// 不更新状态
	}

	if ctx.Value(types.Trigger) == types.CRON_JOB {
		UpdateStatusWithMessage(FetchUpdateBegin, "触发更新")
		defer func() {
			if err == nil && release != nil {

				if !r.ShouldUpgrade(*release, r.SysRoot) {
					UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
					return
				} else {
					if r.IsUpgradable(*release, r.SysRoot) {
						UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
					} else {
						UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
					}
				}
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
		// 如果是HTTP请求的话，则不更新状态
		defer func() {
			if err == nil && release != nil {
				if !r.ShouldUpgrade(*release, r.SysRoot) {
					UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
					return
				} else {
					if r.IsUpgradable(*release, r.SysRoot) {
						UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
					} else {
						UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
					}
				}
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.INSTALL {
		// 如果是HTTP请求的话，则不更新状态
		UpdateStatusWithMessage(InstallBegin, "fetching")
	}

	release, err = r.ImplementService.GetRelease(ctx, tag)
	if err != nil {
		fmt.Println(err)
	}
	return release, err
}

func (r *StatusService) Launch(sysRoot string) error {
	// 在这里会把状态更新为installing或者继续idle
	UpdateStatusWithMessage(InstallBegin, "migration") // 事实上已经没有migration了，但是为了兼容性， 先留着
	defer UpdateStatusWithMessage(InstallBegin, "other")
	// defer UpdateStatusWithMessage(InstallEnd, "migration")
	return r.ImplementService.Launch(sysRoot)
}

func (r *StatusService) VerifyRelease(release codegen.Release) (string, error) {
	return r.ImplementService.VerifyRelease(release)
}

func (r *StatusService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	err := error(nil)
	if ctx.Value(types.Trigger) == types.CRON_JOB {
		UpdateStatusWithMessage(DownloadBegin, "下载中")
		defer func() {
			if err == nil {
				UpdateStatusWithMessage(DownloadEnd, "ready-to-update")
			} else {
				UpdateStatusWithMessage(DownloadError, err.Error())
			}
		}()

	}

	if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
		UpdateStatusWithMessage(DownloadBegin, "http 触发的下载")
		defer func() {
			if err == nil {
				UpdateStatusWithMessage(DownloadEnd, "ready-to-update")
			} else {
				UpdateStatusWithMessage(DownloadError, err.Error())
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.INSTALL {
		UpdateStatusWithMessage(InstallBegin, "downloading")
		defer func() {
			if err != nil {
				UpdateStatusWithMessage(InstallError, err.Error())
			}
		}()
	}

	result, err := r.ImplementService.DownloadRelease(ctx, release, force)
	return result, err
}

func (r *StatusService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	UpdateStatusWithMessage(InstallBegin, types.DECOMPRESS)
	err := r.ImplementService.ExtractRelease(packageFilepath, release)
	defer func() {
		if err != nil {
			UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) PostInstall(release codegen.Release, sysRoot string) error {
	UpdateStatusWithMessage(InstallBegin, types.RESTARTING)
	err := r.ImplementService.PostInstall(release, sysRoot)
	defer func() {
		if err != nil {
			UpdateStatusWithMessage(InstallError, err.Error())
		} else {
			fmt.Println(err)
		}
	}()
	return err
}

func (r *StatusService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return r.ImplementService.ShouldUpgrade(release, sysRoot)
}

func (r *StatusService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	return r.ImplementService.IsUpgradable(release, sysRootPath)
}

func (r *StatusService) PostMigration(sysRoot string) error {
	UpdateStatusWithMessage(InstallBegin, "other")
	err := r.ImplementService.PostMigration(sysRoot)
	defer func() {
		if err == nil {
			UpdateStatusWithMessage(InstallEnd, "up-to-date")
		} else {
			UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) Cronjob(sysRoot string) error {
	return nil
}
