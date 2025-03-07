package service

import (
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/command"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

func SetBetaSubscriptionStatus(status codegen.SetBetaSubscriptionStatusParams) error {
	switch status.Status {
	case codegen.Enable:
		result, err := command.ExecResultStr("channel-tool public")
		if err != nil {
			return err
		}

		if strings.TrimSpace(result) == "Public Test Channel" {
			return nil
		}

	case codegen.Disable:
		result, err := command.ExecResultStr("channel-tool stable")
		if err != nil {
			return err
		}

		if strings.TrimSpace(result) != "Stable Channel" {
			return nil
		}

	}

	return nil
}

func GetBetaSubscriptionStatus() (codegen.Beta, error) {
	stats := InstallerService.Stats()

	if stats.Channel == StableChannelType {
		return codegen.Beta{
			Status: codegen.Disabled,
		}, nil
	}

	return codegen.Beta{
		Status: codegen.Enabled,
	}, nil
}
