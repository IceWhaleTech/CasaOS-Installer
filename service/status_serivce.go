package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
)

type StatusService struct {
	ImplementService                 UpdaterServiceInterface
	SysRoot                          string
	have_other_get_release_flag      bool
	Have_other_get_release_flag_lock sync.RWMutex
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

func (r *StatusService) postGetRelease(ctx context.Context, release *codegen.Release) {
	defer func() {
		r.Have_other_get_release_flag_lock.Lock()
		r.have_other_get_release_flag = false
		r.Have_other_get_release_flag_lock.Unlock()
	}()

	status, _ := GetStatus()
	if status.Status == codegen.Downloading {
		return
	}
	if status.Status == codegen.Installing {
		return
	}

	// 这里怎么判断如果有其它fetching就不搞这个了?
	if !r.ShouldUpgrade(*release, r.SysRoot) {
		UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
		return
	} else {
		if r.IsUpgradable(*release, r.SysRoot) {
			UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
		} else {
			UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
			// 这里应该去触发下载
			go r.DownloadRelease(ctx, *release, false)
		}
		return
	}
}

func (r *StatusService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	// 只允许一个release进 postGetRelease (这个是为了防止多个请求同时触发checksum)(后面已经对checksum做了缓存)，可以考虑不要
	flag := false
	r.Have_other_get_release_flag_lock.Lock()
	if !r.have_other_get_release_flag {
		if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
			r.have_other_get_release_flag = true
			flag = true
		}
	}
	r.Have_other_get_release_flag_lock.Unlock()

	release := &codegen.Release{}

	err := error(nil)
	// 因为更新完进入主页又要拿一次release
	if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
		defer func() {
			if err == nil && release != nil && flag {
				go func() {
					r.postGetRelease(ctx, release)
				}()
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.CRON_JOB {
		UpdateStatusWithMessage(FetchUpdateBegin, "触发更新")

		// 如果是HTTP请求的话，则不更新状态
		defer func() {
			if err == nil && release != nil {
				go func() {
					r.postGetRelease(ctx, release)
				}()
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

	local_status, _ := GetStatus()
	if local_status.Status == codegen.Downloading {
		return "", fmt.Errorf("downloading")
	}
	if local_status.Status == codegen.FetchUpdating {
		return "", fmt.Errorf("fecthing")
	}
	if local_status.Status == codegen.Installing && ctx.Value(types.Trigger) != types.INSTALL {
		return "", fmt.Errorf("installing")
	}

	if ctx.Value(types.Trigger) == types.CRON_JOB {
		fmt.Println("开始下载的状态")

		UpdateStatusWithMessage(DownloadBegin, "下载中")
		defer func() {
			fmt.Println("下载完成")

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

func (r *StatusService) InstallInfo(release codegen.Release, sysRootPath string) (string, error) {
	return r.ImplementService.InstallInfo(release, sysRootPath)
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
