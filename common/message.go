package common

import (
	"github.com/IceWhaleTech/CasaOS-Common/utils"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
)

var EventTypes = []message_bus.EventType{
	// check update
	EventTypeCheckUpdateBegin, EventTypeCheckUpdateEnd, EventTypeCheckUpdateError,

	// download update
	EventTypeDownloadUpdateBegin, EventTypeDownloadUpdateEnd, EventTypeDownloadUpdateError,

	// install update
	EventTypeInstallUpdateBegin, EventTypeInstallUpdateEnd, EventTypeInstallUpdateError,
}

var (
	PropertyTypeAppName = message_bus.PropertyType{
		Name:        "app:name",
		Description: utils.Ptr("name of the app which could be a container image name including version, a snap name or the name of any other forms of app"),
		Example:     utils.Ptr("hello-world:latest (this is the name of a container image"),
	}
)

var (
	PropertyTypeMessage = message_bus.PropertyType{
		Name:        "message",
		Description: utils.Ptr("message at different levels, typically for error"),
	}
)

var (
	EventTypeCheckUpdateBegin = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:check-update-begin",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeCheckUpdateEnd = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:check-update-end",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeCheckUpdateError = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:check-update-error",
		PropertyTypeList: []message_bus.PropertyType{},
	}

	EventTypeDownloadUpdateBegin = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:download-update-begin",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeDownloadUpdateEnd = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:download-update-end",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeDownloadUpdateError = message_bus.EventType{
		SourceID: InstallerServiceName,
		Name:     "installer:download-update-error",
		PropertyTypeList: []message_bus.PropertyType{
			PropertyTypeAppName,
			PropertyTypeMessage,
		},
	}

	EventTypeInstallUpdateBegin = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:install-update-begin",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeInstallUpdateEnd = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:install-update-end",
		PropertyTypeList: []message_bus.PropertyType{},
	}
	EventTypeInstallUpdateError = message_bus.EventType{
		SourceID:         InstallerServiceName,
		Name:             "installer:install-update-error",
		PropertyTypeList: []message_bus.PropertyType{},
	}
)
