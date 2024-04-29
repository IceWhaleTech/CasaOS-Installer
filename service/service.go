package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"go.uber.org/zap"
)

type EventType string

var MyService Services
var InstallerService UpdaterServiceInterface

type Services interface {
	Gateway() (external.ManagementService, error)
	MessageBus() (*message_bus.ClientWithResponses, error)
}

type services struct {
	runtimePath string
}

func NewService(RuntimePath string) Services {
	// UpdateStatus(codegen.Status{
	// 	Status: codegen.Idle,
	// })
	return &services{
		runtimePath: RuntimePath,
	}
}

func NewInstallerService(sysRoot string) UpdaterServiceInterface {

	CleanupOfflineRAUCTemp(sysRoot)

	installMethod, err := GetInstallMethod(sysRoot)
	if err != nil {
		logger.Error("failed to get install method", zap.Error(err))
	}

	if installMethod == RAUC {
		logger.Info("RAUC Online mode")
		return &RAUCService{
			InstallRAUCHandler: InstallRAUCImp,
			DownloadHandler:    nil,
			CheckSumHandler:    checksum.OnlineRaucChecksumExist,
			UrlHandler:         HyperFileTagReleaseUrl,
		}
	}

	if installMethod == RAUCOFFLINE {
		logger.Info("RAUC Offline mode")
		return &RAUCOfflineService{
			SysRoot:            sysRoot,
			InstallRAUCHandler: InstallRAUCImp,
			CheckSumHandler:    checksum.OfflineTarExist,
			GetRAUCInfo:        GetRAUCInfo,
		}
	}

	// if installMethod == TAR {
	// 	return &RAUCService{
	// 		InstallRAUCHandler: InstallRAUCImp,
	// 		CheckSumHandler:    checksum.OnlineRAUCExist,
	// 		UrlHandler:         HyperFileTagReleaseUrl,
	// 	}
	// }
	logger.Info("default mode")
	return &RAUCService{
		InstallRAUCHandler: InstallRAUCImp,
		DownloadHandler:    nil,
		CheckSumHandler:    checksum.OnlineRaucChecksumExist,
		UrlHandler:         HyperFileTagReleaseUrl,
	}
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
		fmt.Println("Warning: failed to publish event - message bus service didn't running")
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
		logger.Error("failed to publish event", zap.String("status code", response.Status()), zap.String("body", string(response.Body)))
	}
}
