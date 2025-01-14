package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

var versionRegexp = regexp.MustCompile(`^v(\d)+\.(\d)+.(\d)+(-(alpha|beta)?(\d)+)?$`)

func NewStatusService(implementService UpdaterServiceInterface, sysRoot string) *StatusService {
	statusService := &StatusService{
		ImplementService: implementService,
		SysRoot:          sysRoot,
	}
	statusService.status = codegen.Status{
		Status: codegen.Idle,
	}

	//go:nocheckrace
	go func() {
		release, err := implementService.GetRelease(context.Background(), GetReleaseBranch(sysRoot), true)
		if err != nil {
			logger.Error("fail get release", zap.Error(err))
		} else {
			statusService.release = release
		}
	}()
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

func (r *StatusService) GetRelease(ctx context.Context, tag string, useCache bool) (*codegen.Release, error) {
	if r.release == nil {
		release, err := r.ImplementService.GetRelease(ctx, tag, true)
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
	r.UpdateStatusWithMessage(InstallBegin, types.OTHER)
	err := r.ImplementService.PostMigration(sysRoot)
	defer func() {
		if err == nil {
			r.UpdateStatusWithMessage(InstallEnd, types.UP_TO_DATE)
		} else {
			r.UpdateStatusWithMessage(InstallError, err.Error())
		}
	}()
	return err
}

func (r *StatusService) CleanUpOldRelease(sysRoot string) error {
	currentVersion, err := CurrentReleaseVersion(sysRoot)
	if err != nil {
		logger.Error("error when trying to get current release version", zap.Error(err))
		return err
	}

	dirs, err := filepath.Glob(filepath.Join(sysRoot, "DATA", "rauc", "releases", "*"))
	if err != nil {
		logger.Error("error when trying to get all dirs in release", zap.Error(err))
		return err
	}

	for _, dir := range dirs {
		baseDir := filepath.Base(dir)
		if !versionRegexp.MatchString(baseDir) {
			continue
		}

		version := strings.TrimPrefix(baseDir, "v")
		var whiteList []string

		if IsNewerVersionString(currentVersion.String(), version) {
			logger.Info("newer version found, skip clean up", zap.String("dir", dir))
			continue
		} else {
			logger.Info("cleanning up", zap.String("dir", dir))
			whiteList = []string{"zimaos_zimacube-" + version + ".raucb", "checksums.txt"}

			//! Important!: 这里不能删除，为了当前版本在重启以后还能看到更新日志弹框
			if !(currentVersion.String() == version) {
				whiteList = append(whiteList, "release.yaml")
			}

			if err := internal.CleanWithWhiteList(dir, whiteList); err != nil {
				logger.Error("error when trying to clean up release", zap.Error(err))
			}
		}
	}

	return nil
}

func (r *StatusService) Cronjob(ctx context.Context, sysRoot string) error {
	logger.Info("start a check update job")

	ctx = context.WithValue(ctx, types.Trigger, types.CRON_JOB)

	status, _ := r.GetStatus()
	if status.Status == codegen.Downloading {
		logger.Info("downloading, skip")
		return nil
	}

	if status.Status == codegen.Installing {
		logger.Info("installing, skip")
		return nil
	}

	r.UpdateStatusWithMessage(FetchUpdateBegin, types.FETCHING)
	logger.Info("start to fetch  release ", zap.Any("info", r.Stats()), zap.Any("array", config.ServerInfo.Mirrors))

	release, err := r.ImplementService.GetRelease(ctx, GetReleaseBranch(sysRoot), false)
	if err != nil {
		r.UpdateStatusWithMessage(FetchUpdateError, err.Error())
		logger.Error("error when trying to get release", zap.Error(err))
		return err
	}
	r.release = release
	r.UpdateStatusWithMessage(FetchUpdateEnd, types.OUT_OF_DATE)

	logger.Info("get release success", zap.String("release version", release.Version))

	// cache release packages if not already cached
	shouldUpgrade := r.ShouldUpgrade(*release, sysRoot)

	releaseFilePath := ""

	if shouldUpgrade {
		if release.Background == nil {
			logger.Error("release.Background is nil", zap.Any("info", r.Stats()))
		} else {
			go internal.DownloadReleaseBackground(*release.Background, release.Version)
		}

		releaseFilePath, err = r.DownloadRelease(ctx, *release, true)
		if err != nil {
			logger.Error("error when trying to download release", zap.Error(err), zap.String("release file path", releaseFilePath), zap.Any("info", r.Stats()))
			r.UpdateStatusWithMessage(DownloadError, err.Error())
		} else {
			logger.Info("download release rauc update package success")
			r.UpdateStatusWithMessage(DownloadEnd, types.READY_TO_UPDATE)
		}
	} else {
		releaseFilePath, err = r.InstallInfo(*release, sysRoot)
		if err != nil {
			logger.Error("error when trying to get install info", zap.Error(err))
		}

		logger.Info("system is up to date", zap.Any("info", r.Stats()))
		r.UpdateStatusWithMessage(FetchUpdateEnd, types.UP_TO_DATE)
	}

	if releaseFilePath == "" {
		if shouldUpgrade {
			logger.Error("release file path is empty")
		}
	} else {
		releaseDir := filepath.Dir(releaseFilePath)
		latestReleaseDir := filepath.Join(filepath.Dir(releaseDir), "latest")

		os.Remove(latestReleaseDir)
		err = os.Symlink(releaseDir, latestReleaseDir)

		logger.Info("create latest symlink ok", zap.Error(err), zap.Any("releaseDir", releaseDir), zap.Any("latestReleaseDir", latestReleaseDir))

		if err = r.CleanUpOldRelease(sysRoot); err != nil {
			logger.Error("error when trying to clean up", zap.Error(err))
		}
	}

	return nil
}

func (r *StatusService) Stats() UpdateServerStats {
	return r.ImplementService.Stats()
}
