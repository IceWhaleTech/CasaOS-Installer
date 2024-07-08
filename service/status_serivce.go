package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"go.uber.org/zap"
)

// TODO: 考虑重构这里。当前前端页面设计的时候是需要后端的具体的状态的，比如正在抓取、正在下载。
// 后面一个状态需要同步给前端和Message Bus，然后一个语法是ing、一个是done。我加了一个中间层来兼容两边。
// 但是现在业务发生了变化，考虑是不需要重构这里减少复杂性。
type StatusService struct {
	ImplementService UpdaterServiceInterface
	release          *codegen.Release
	SysRoot          string
	status           codegen.Status
	message          string
	lock             sync.RWMutex
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

var EventTypeMapStatus = map[EventType]codegen.Status{
	DownloadBegin: {Status: codegen.Downloading},
	DownloadEnd:   {Status: codegen.Idle},
	DownloadError: {Status: codegen.Idle},

	FetchUpdateBegin: {Status: codegen.FetchUpdating},
	FetchUpdateEnd:   {Status: codegen.Idle},
	FetchUpdateError: {Status: codegen.Idle},

	InstallBegin: {Status: codegen.Installing},
	InstallEnd:   {Status: codegen.Idle},
	InstallError: {Status: codegen.InstallError},
}

var EventTypeMapMessageType = map[EventType]message_bus.EventType{
	FetchUpdateBegin: common.EventTypeCheckUpdateBegin,
	FetchUpdateEnd:   common.EventTypeCheckUpdateEnd,
	FetchUpdateError: common.EventTypeCheckUpdateError,

	DownloadBegin: common.EventTypeDownloadUpdateBegin,
	DownloadEnd:   common.EventTypeDownloadUpdateEnd,
	DownloadError: common.EventTypeDownloadUpdateError,

	InstallBegin: common.EventTypeInstallUpdateBegin,
	InstallEnd:   common.EventTypeInstallUpdateEnd,
	InstallError: common.EventTypeInstallUpdateError,
}

func NewStatusService(implementService UpdaterServiceInterface, sysRoot string) *StatusService {
	statusService := &StatusService{
		ImplementService: implementService,
		SysRoot:          sysRoot,
	}
	statusService.status = codegen.Status{
		Status: codegen.Idle,
	}
	return statusService
}

func (r *StatusService) GetStatus() (codegen.Status, string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.status, r.message
}

func (r *StatusService) UpdateStatusWithMessage(eventType EventType, eventMessage string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	switch eventType {
	case DownloadBegin:
		r.status = EventTypeMapStatus[DownloadBegin]
	case DownloadEnd:
		r.status = EventTypeMapStatus[DownloadEnd]
	case DownloadError:
		r.status = EventTypeMapStatus[DownloadError]
	case FetchUpdateBegin:
		r.status = EventTypeMapStatus[FetchUpdateBegin]
	case FetchUpdateEnd:
		r.status = EventTypeMapStatus[FetchUpdateEnd]
	case FetchUpdateError:
		r.status = EventTypeMapStatus[FetchUpdateError]
	case InstallBegin:
		r.status = EventTypeMapStatus[InstallBegin]
	case InstallEnd:
		r.status = EventTypeMapStatus[InstallEnd]
	case InstallError:
		r.status = EventTypeMapStatus[InstallError]
	default:
		r.status = codegen.Status{
			Status: codegen.Idle,
		}
	}

	r.message = eventMessage

	ctx := context.Background()

	event := EventTypeMapMessageType[eventType]

	go PublishEventWrapper(ctx, event, map[string]string{
		common.PropertyTypeMessage.Name: eventMessage,
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

func (r *StatusService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	// TODO: cache release in disk
	if r.release == nil {
		release, err := r.ImplementService.GetRelease(ctx, tag)
		if err != nil {
			return nil, err
		}
		r.release = release
	}
	return r.release, nil
}

func (r *StatusService) Launch(sysRoot string) error {
	// 事实上已经没有migration了，但是为了兼容性， 先留着
	r.UpdateStatusWithMessage(InstallBegin, types.MIGRATION)
	defer r.UpdateStatusWithMessage(InstallBegin, types.OTHER)
	return nil
}

func (r *StatusService) VerifyRelease(release codegen.Release) (string, error) {
	return r.ImplementService.VerifyRelease(release)
}

func (r *StatusService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	err := error(nil)

	localStatus, _ := r.GetStatus()
	if localStatus.Status == codegen.Downloading {
		return "", fmt.Errorf("downloading")
	}
	if localStatus.Status == codegen.Installing && ctx.Value(types.Trigger) != types.INSTALL {
		return "", fmt.Errorf("installing")
	}

	switch ctx.Value(types.Trigger) {
	case types.CRON_JOB:
		r.UpdateStatusWithMessage(DownloadBegin, types.DOWNLOADING)
		defer func() {
			if err == nil {
				r.UpdateStatusWithMessage(DownloadEnd, types.READY_TO_UPDATE)
			} else {
				r.UpdateStatusWithMessage(DownloadError, err.Error())
			}
		}()

	case types.INSTALL:
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
	return nil
}

func (r *StatusService) PostInstall(release codegen.Release, sysRoot string) error {
	r.UpdateStatusWithMessage(InstallBegin, types.RESTARTING)
	err := r.ImplementService.PostInstall(release, sysRoot)
	defer func() {
		if err != nil {
			r.UpdateStatusWithMessage(InstallError, err.Error())
		} else {
			logger.Error("error when trying to post install", zap.Error(err))
		}
	}()
	return err
}

func (r *StatusService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	su := r.ImplementService.ShouldUpgrade(release, sysRoot)
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

func (r *StatusService) Cronjob(ctx context.Context, sysRoot string) error {
	logger.Info("start a check update job")

	ctx = context.WithValue(ctx, types.Trigger, types.CRON_JOB)

	status, _ := r.GetStatus()
	if status.Status == codegen.Downloading {
		return nil
	}

	if status.Status == codegen.Installing {
		return nil
	}

	r.UpdateStatusWithMessage(FetchUpdateBegin, types.FETCHING)
	logger.Info("start to fetch online release ", zap.Any("array", config.ServerInfo.Mirrors))

	release, err := r.ImplementService.GetRelease(ctx, GetReleaseBranch(sysRoot))
	if err != nil {
		r.UpdateStatusWithMessage(FetchUpdateError, err.Error())
		logger.Error("error when trying to get release", zap.Error(err))
		return err
	}
	r.release = release

	logger.Info("get online release success", zap.String("online release version", release.Version))

	r.UpdateStatusWithMessage(DownloadBegin, types.DOWNLOADING)
	if release.Background == nil {
		logger.Error("release.Background is nil")
	} else {
		go internal.DownloadReleaseBackground(*release.Background, release.Version)
	}

	// cache release packages if not already cached
	shouldUpgrade := r.ShouldUpgrade(*release, sysRoot)

	if shouldUpgrade {
		r.UpdateStatusWithMessage(FetchUpdateEnd, types.OUT_OF_DATE)

		releaseFilePath, err := r.DownloadRelease(ctx, *release, true)
		logger.Info("download release rauc update package success")

		if err != nil {
			logger.Error("error when trying to download release", zap.Error(err), zap.String("release file path", releaseFilePath))
			r.UpdateStatusWithMessage(DownloadError, err.Error())
		} else {
			logger.Info("system is ready to update")
			r.UpdateStatusWithMessage(DownloadEnd, types.READY_TO_UPDATE)
		}
	} else {
		logger.Info("system is up to date")
		r.UpdateStatusWithMessage(FetchUpdateEnd, types.UP_TO_DATE)
	}

	return nil
}
