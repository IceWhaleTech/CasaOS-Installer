package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
)

type StatusService struct {
	ImplementService                 UpdaterServiceInterface
	SysRoot                          string
	have_other_get_release_flag      bool
	Have_other_get_release_flag_lock sync.RWMutex

	status  codegen.Status
	message string
	lock    sync.RWMutex

	EventTypeMapStatus      map[EventType]codegen.Status
	EventTypeMapMessageType map[EventType]message_bus.EventType
}

const (
	DownloadBegin    EventType = "downloadBegin"
	DownloadEnd      EventType = "downloadEnd"
	DownloadError    EventType = "downloadError"
	FetchUpdateEnd   EventType = "fetchUpdateEnd"
	FetchUpdateBegin EventType = "fetchUpdateBegin"
	FetchUpdateError EventType = "fetchUpdateError"

	Idle         EventType = "idle"
	InstallEnd   EventType = "installEnd"
	InstallBegin EventType = "installBegin"
	InstallError EventType = "installError"
)

func (r *StatusService) InitEventTypeMapStatus() {
	r.EventTypeMapStatus = make(map[EventType]codegen.Status)
	r.EventTypeMapMessageType = make(map[EventType]message_bus.EventType)

	r.EventTypeMapStatus[DownloadBegin] = codegen.Status{
		Status: codegen.Downloading,
	}
	r.EventTypeMapStatus[DownloadEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	r.EventTypeMapStatus[DownloadError] = codegen.Status{
		Status: codegen.DownloadError,
	}

	r.EventTypeMapStatus[FetchUpdateBegin] = codegen.Status{
		Status: codegen.FetchUpdating,
	}
	r.EventTypeMapStatus[FetchUpdateEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	r.EventTypeMapStatus[FetchUpdateError] = codegen.Status{
		Status: codegen.FetchError,
	}

	r.EventTypeMapStatus[InstallBegin] = codegen.Status{
		Status: codegen.Installing,
	}
	r.EventTypeMapStatus[InstallEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	r.EventTypeMapStatus[InstallError] = codegen.Status{
		Status: codegen.InstallError,
	}

	r.EventTypeMapMessageType[FetchUpdateBegin] = common.EventTypeCheckUpdateBegin
	r.EventTypeMapMessageType[FetchUpdateEnd] = common.EventTypeCheckUpdateEnd
	r.EventTypeMapMessageType[FetchUpdateError] = common.EventTypeCheckUpdateError

	r.EventTypeMapMessageType[DownloadBegin] = common.EventTypeDownloadUpdateBegin
	r.EventTypeMapMessageType[DownloadEnd] = common.EventTypeDownloadUpdateEnd
	r.EventTypeMapMessageType[DownloadError] = common.EventTypeDownloadUpdateError

	r.EventTypeMapMessageType[InstallBegin] = common.EventTypeInstallUpdateBegin
	r.EventTypeMapMessageType[InstallEnd] = common.EventTypeInstallUpdateEnd
	r.EventTypeMapMessageType[InstallError] = common.EventTypeInstallUpdateError
}

func NewStatusService(implementService UpdaterServiceInterface, sysRoot string) *StatusService {
	statusService := &StatusService{
		ImplementService:                 implementService,
		SysRoot:                          sysRoot,
		Have_other_get_release_flag_lock: sync.RWMutex{},
	}
	statusService.InitEventTypeMapStatus()
	return statusService
}

func (r *StatusService) GetStatus() (codegen.Status, string) {
	return r.status, r.message
}

func (r *StatusService) UpdateStatusWithMessage(eventType EventType, newPackageStatus string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if (eventType != InstallEnd && eventType != InstallError && eventType != InstallBegin) && (r.status.Status == codegen.Installing) {
		return
	}

	switch eventType {
	case DownloadBegin:
		r.status = r.EventTypeMapStatus[DownloadBegin]
	case DownloadEnd:
		r.status = r.EventTypeMapStatus[DownloadEnd]
	case DownloadError:
		r.status = r.EventTypeMapStatus[DownloadError]
	case FetchUpdateBegin:
		r.status = r.EventTypeMapStatus[FetchUpdateBegin]
	case FetchUpdateEnd:
		r.status = r.EventTypeMapStatus[FetchUpdateEnd]
	case FetchUpdateError:
		r.status = r.EventTypeMapStatus[FetchUpdateError]
	case InstallBegin:
		r.status = r.EventTypeMapStatus[InstallBegin]
	case InstallEnd:
		r.status = r.EventTypeMapStatus[InstallEnd]
	case InstallError:
		r.status = r.EventTypeMapStatus[InstallError]
	}

	r.message = newPackageStatus

	ctx := context.Background()

	// 这里怎么map一下?🤔
	event := r.EventTypeMapMessageType[eventType]

	go PublishEventWrapper(ctx, event, map[string]string{
		common.PropertyTypeMessage.Name: newPackageStatus,
	})
}

func (r *StatusService) Install(release codegen.Release, sysRoot string) error {
	r.UpdateStatusWithMessage(InstallBegin, types.INSTALLING)
	err := r.ImplementService.Install(release, sysRoot)
	defer func() {
		if err != nil {
			r.UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) postGetRelease(ctx context.Context, release *codegen.Release) {
	// defer func() {
	// 	r.Have_other_get_release_flag_lock.Lock()
	// 	r.have_other_get_release_flag = false
	// 	r.Have_other_get_release_flag_lock.Unlock()
	// }()

	status, _ := r.GetStatus()
	if status.Status == codegen.Downloading {
		return
	}
	if status.Status == codegen.Installing {
		return
	}

	// 这里怎么判断如果有其它fetching就不搞这个了?
	if !r.ShouldUpgrade(*release, r.SysRoot) {
		r.UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
		return
	} else {
		if r.IsUpgradable(*release, r.SysRoot) {
			r.UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
		} else {
			r.UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
			// 这里应该去触发下载
			go r.DownloadRelease(ctx, *release, false)
		}
		return
	}
}

func (r *StatusService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	// // 只允许一个release进 postGetRelease (这个是为了防止多个请求同时触发checksum)(后面已经对checksum做了缓存)，可以考虑不要
	// flag := false
	// r.Have_other_get_release_flag_lock.Lock()
	// if !r.have_other_get_release_flag {
	// 	if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
	// 		r.have_other_get_release_flag = true
	// 		flag = true
	// 	}
	// }
	// r.Have_other_get_release_flag_lock.Unlock()

	if ctx.Value(types.Trigger) == types.CRON_JOB {
		r.UpdateStatusWithMessage(FetchUpdateBegin, "fetching")
	}

	if ctx.Value(types.Trigger) == types.INSTALL {
		// 如果是HTTP请求的话，则不更新状态
		r.UpdateStatusWithMessage(InstallBegin, types.FETCHING)
	}

	var release = &codegen.Release{}
	err := error(nil)
	release, err = r.ImplementService.GetRelease(ctx, tag)
	if err != nil {
		logger.Error(fmt.Sprintf("Get Release Faile %s tag:%s", err.Error(), tag))
	} else {
		logger.Info(fmt.Sprintf("Get Release success! %s", release.Version))
	}

	// // 因为更新完进入主页又要拿一次release
	// if ctx.Value(types.Trigger) == types.HTTP_REQUEST || ctx.Value(types.Trigger) == types.CRON_JOB {
	// 	defer func() {
	// 		if err == nil && release != nil {
	// 			go func() {
	// 				r.postGetRelease(ctx, release)
	// 			}()
	// 		}
	// 	}()
	// }

	// if {
	// 	UpdateStatusWithMessage(FetchUpdateBegin, "fetching")

	// 	// 如果是HTTP请求的话，则不更新状态
	// 	defer func() {
	// 		if err == nil && release != nil {
	// 			go func() {
	r.postGetRelease(ctx, release)
	// 			}()
	// 		}
	// 	}()
	// }

	return release, err
}

func (r *StatusService) Launch(sysRoot string) error {
	// 在这里会把状态更新为installing或者继续idle
	r.UpdateStatusWithMessage(InstallBegin, "migration") // 事实上已经没有migration了，但是为了兼容性， 先留着
	defer r.UpdateStatusWithMessage(InstallBegin, "other")
	// defer UpdateStatusWithMessage(InstallEnd, "migration")
	//return r.ImplementService.Launch(sysRoot)
	return nil
}

func (r *StatusService) VerifyRelease(release codegen.Release) (string, error) {
	return r.ImplementService.VerifyRelease(release)
}

func (r *StatusService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	err := error(nil)

	local_status, _ := r.GetStatus()
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

		r.UpdateStatusWithMessage(DownloadBegin, types.DOWNLOADING)
		defer func() {
			if err == nil {
				r.UpdateStatusWithMessage(DownloadEnd, types.READY_TO_UPDATE)
			} else {
				r.UpdateStatusWithMessage(DownloadError, err.Error())
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.HTTP_REQUEST {
		r.UpdateStatusWithMessage(DownloadBegin, "http 触发的下载")
		defer func() {
			if err == nil {
				r.UpdateStatusWithMessage(DownloadEnd, types.READY_TO_UPDATE)
			} else {
				r.UpdateStatusWithMessage(DownloadError, err.Error())
			}
		}()
	}

	if ctx.Value(types.Trigger) == types.INSTALL {
		r.UpdateStatusWithMessage(InstallBegin, types.DOWNLOADING)
		defer func() {
			if err != nil {
				r.UpdateStatusWithMessage(InstallError, err.Error())
			}
		}()
	}

	result, err := r.ImplementService.DownloadRelease(ctx, release, force)
	return result, err
}

func (r *StatusService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	r.UpdateStatusWithMessage(InstallBegin, types.DECOMPRESS)
	// err := r.ImplementService.ExtractRelease(packageFilepath, release)
	// defer func() {
	// 	if err != nil {
	// 		UpdateStatusWithMessage(InstallError, err.Error())
	// 	}
	// }()
	//return err
	return nil
}

func (r *StatusService) PostInstall(release codegen.Release, sysRoot string) error {
	r.UpdateStatusWithMessage(InstallBegin, types.RESTARTING)
	err := r.ImplementService.PostInstall(release, sysRoot)
	defer func() {
		if err != nil {
			r.UpdateStatusWithMessage(InstallError, err.Error())
		} else {
			fmt.Println(err)
		}
	}()
	return err
}

func (r *StatusService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {

	su := r.ImplementService.ShouldUpgrade(release, sysRoot)

	if !su {
		r.UpdateStatusWithMessage(FetchUpdateEnd, "up-to-date")
	} else {
		if r.IsUpgradable(release, r.SysRoot) {
			r.UpdateStatusWithMessage(FetchUpdateEnd, "ready-to-update")
		} else {
			r.UpdateStatusWithMessage(FetchUpdateEnd, "out-of-date")
		}
	}
	return su
}

func (r *StatusService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	return r.ImplementService.IsUpgradable(release, sysRootPath)
}

func (r *StatusService) InstallInfo(release codegen.Release, sysRootPath string) (string, error) {
	return r.ImplementService.InstallInfo(release, sysRootPath)
}

func (r *StatusService) PostMigration(sysRoot string) error {
	r.UpdateStatusWithMessage(InstallBegin, "other")
	err := r.ImplementService.PostMigration(sysRoot)
	defer func() {
		if err == nil {
			r.UpdateStatusWithMessage(InstallEnd, "up-to-date")
		} else {
			r.UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) Cronjob(sysRoot string) error {
	return nil
}
