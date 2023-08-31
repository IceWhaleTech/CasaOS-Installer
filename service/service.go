package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"go.uber.org/zap"
)

type EventType string

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

var EventTypeMapStatus = make(map[EventType]codegen.Status)
var EventTypeMapMessageType = make(map[EventType]message_bus.EventType)

var MyService Services
var InstallerService InstallerServices

// TODO move to another place
var status codegen.Status
var packageStatus string
var lock sync.RWMutex

type Services interface {
	Gateway() (external.ManagementService, error)
	MessageBus() (*message_bus.ClientWithResponses, error)
}

type InstallerServices interface {
	GetRelease(ctx context.Context, tag string) (*codegen.Release, error)
	VerifyRelease(release codegen.Release) (string, error)
	DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error)
	ExtractRelease(packageFilepath string, release codegen.Release) error
	GetMigrationInfo(ctx context.Context, release codegen.Release) error
	DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error
	Install(release codegen.Release, sysRoot string) error
	MigrationInLaunch(sysRoot string) error
}

type services struct {
	runtimePath string
}

// func UpdateStatus(newStatus codegen.Status) {
// 	// if status move to server. the function can be reuse.
// 	lock.Lock()
// 	defer lock.Unlock()

// 	status = newStatus
// 	message = ""

// 	// 把消息总线和状态放在一起。
// 	ctx := context.Background()
// 	go PublishEventWrapper(ctx, common.EventTypeDownloadUpdateEnd, nil)
// }

func GetStatus() (codegen.Status, string) {
	lock.RLock()
	defer lock.RUnlock()
	return status, packageStatus
}

func UpdateStatusWithMessage(eventType, newPackageStatus string) {
	lock.Lock()
	defer lock.Unlock()

	// TODO only run once
	InitEventTypeMapStatus()

	switch eventType {
	case string(DownloadBegin):
		status = EventTypeMapStatus[DownloadBegin]
	case string(DownloadEnd):
		status = EventTypeMapStatus[DownloadEnd]
	case string(FetchUpdateBegin):
		status = EventTypeMapStatus[FetchUpdateBegin]
	case string(FetchUpdateEnd):
		status = EventTypeMapStatus[FetchUpdateEnd]
	case string(InstallBegin):
		status = EventTypeMapStatus[InstallBegin]
	case string(InstallEnd):
		status = EventTypeMapStatus[InstallEnd]
	case string(InstallError):
		status = EventTypeMapStatus[InstallError]
	}

	packageStatus = newPackageStatus

	ctx := context.Background()

	// 这里怎么map一下?🤔
	event := EventTypeMapMessageType[EventType(eventType)]

	go PublishEventWrapper(ctx, event, map[string]string{
		common.PropertyTypeMessage.Name: newPackageStatus,
	})
}

func InitEventTypeMapStatus() {
	EventTypeMapStatus[DownloadBegin] = codegen.Status{
		Status: codegen.Downloading,
	}
	EventTypeMapStatus[DownloadEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	EventTypeMapStatus[DownloadError] = codegen.Status{
		Status: codegen.Idle,
	}

	EventTypeMapStatus[FetchUpdateBegin] = codegen.Status{
		Status: codegen.FetchUpdating,
	}
	EventTypeMapStatus[FetchUpdateEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	EventTypeMapStatus[FetchUpdateError] = codegen.Status{
		Status: codegen.Idle,
	}

	EventTypeMapStatus[InstallBegin] = codegen.Status{
		Status: codegen.Installing,
	}
	EventTypeMapStatus[InstallEnd] = codegen.Status{
		Status: codegen.Idle,
	}
	EventTypeMapStatus[InstallError] = codegen.Status{
		Status: codegen.Idle,
	}

	EventTypeMapMessageType[FetchUpdateBegin] = common.EventTypeCheckUpdateBegin
	EventTypeMapMessageType[FetchUpdateEnd] = common.EventTypeCheckUpdateEnd
	EventTypeMapMessageType[FetchUpdateError] = common.EventTypeCheckUpdateError

	EventTypeMapMessageType[DownloadBegin] = common.EventTypeDownloadUpdateBegin
	EventTypeMapMessageType[DownloadEnd] = common.EventTypeDownloadUpdateEnd
	EventTypeMapMessageType[DownloadError] = common.EventTypeDownloadUpdateError

	EventTypeMapMessageType[InstallBegin] = common.EventTypeInstallUpdateBegin
	EventTypeMapMessageType[InstallEnd] = common.EventTypeInstallUpdateEnd
	EventTypeMapMessageType[InstallError] = common.EventTypeInstallUpdateError
}

func NewService(RuntimePath string) Services {
	// UpdateStatus(codegen.Status{
	// 	Status: codegen.Idle,
	// })
	return &services{
		runtimePath: RuntimePath,
	}
}

func NewInstallerService(sysRoot string) InstallerServices {
	installMethod, err := GetInstallMethod(sysRoot)
	if err != nil {
		panic(err)
	}

	// 这里搞个工厂模式。

	if installMethod == "rauc" {
		return &RAUCService{
			InstallRAUCHandler: InstallRAUCHandlerV1,
		}
	}
	// 回头做这个社区版。
	// if installMethod == "tar" {
	// 	return &TarService{}
	// }

	panic(fmt.Errorf("install method %s not supported", installMethod))
}

func (s *services) Gateway() (external.ManagementService, error) {
	return external.NewManagementService(s.runtimePath)
}

func (s *services) MessageBus() (*message_bus.ClientWithResponses, error) {
	return message_bus.NewClientWithResponses("", func(c *message_bus.Client) error {
		// error will never be returned, as we always want to return a client, even with wrong address,
		// in order to avoid panic.
		//
		// If we don't avoid panic, message bus becomes a hard dependency, which is not what we want.

		messageBusAddress, err := external.GetMessageBusAddress(config.CommonInfo.RuntimePath)
		if err != nil {
			c.Server = "message bus address not found"
			return nil
		}

		c.Server = messageBusAddress
		return nil
	})
}

func PublishEventWrapper(ctx context.Context, eventType message_bus.EventType, properties map[string]string) {
	if MyService == nil {
		fmt.Println("Warning: failed to publish event - messsage bus service didn't running")
		return
	}

	messageBus, err := MyService.MessageBus()
	if err != nil {
		logger.Error("failed to publish event", zap.Error(err))
		return
	}

	if properties == nil {
		properties = map[string]string{}
	}

	// merge with properties from context
	for k, v := range common.PropertiesFromContext(ctx) {
		properties[k] = v
	}

	response, err := messageBus.PublishEventWithResponse(ctx, common.InstallerServiceName, eventType.Name, properties)
	if err != nil {
		logger.Error("failed to publish event", zap.Error(err))
		return
	}
	defer response.HTTPResponse.Body.Close()

	if response.StatusCode() != http.StatusOK {
		logger.Error("failed to publish event", zap.String("status code", response.Status()))
	}
}
