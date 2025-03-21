package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/command"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"go.uber.org/zap"
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

		sysRoot := "/media/ZimaOS-HD"

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
			if IsNewerVersionString(currentVersion.String(), version) && strings.Contains(version, "beta") {
				if err := os.RemoveAll(dir); err != nil {
					logger.Error("error when trying to remove dir", zap.Error(err))
				}
			}
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
