package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IceWhaleTech/CasaOS-Common/external"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"go.uber.org/zap"
)

var MyService Services

type Services interface {
	Gateway() (external.ManagementService, error)
	MessageBus() (*message_bus.ClientWithResponses, error)
}

type services struct {
	runtimePath string
}

func NewService(RuntimePath string) Services {
	return &services{
		runtimePath: RuntimePath,
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
