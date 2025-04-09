package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"go.uber.org/zap"
)

func SetBetaSubscriptionStatus(status codegen.SetBetaSubscriptionStatusParams) error {
	switch status.Status {
	case codegen.Enable:
		config.ServerInfo.Mirrors = ChannelData[PublicTestChannelType]
		config.Cfg.Section("server").Key("mirrors").SetValue(strings.Join(config.ServerInfo.Mirrors, ","))
		err := config.Cfg.SaveTo(config.ConfigFilePath)
		if err != nil {
			fmt.Printf("Fail to save file: %v", err)
		}

		go InstallerService.Cronjob(context.Background(), config.SysRoot)
	case codegen.Disable:
		config.ServerInfo.Mirrors = ChannelData[StableChannelType]
		config.Cfg.Section("server").Key("mirrors").SetValue(strings.Join(config.ServerInfo.Mirrors, ","))
		err := config.Cfg.SaveTo(config.ConfigFilePath)
		if err != nil {
			fmt.Printf("Fail to save file: %v", err)
		}

		go func() {
			dirs, err := filepath.Glob(filepath.Join(config.SysRoot, "DATA", "rauc", "releases", "*"))
			if err != nil {
				logger.Error("error when trying to get all dirs in release", zap.Error(err))
				return
			}

			for _, dir := range dirs {
				baseDir := filepath.Base(dir)
				if !versionRegexp.MatchString(baseDir) {
					continue
				}

				version := strings.TrimPrefix(baseDir, "v")
				if strings.Contains(version, "beta") {
					if err := os.RemoveAll(dir); err != nil {
						logger.Error("error when trying to remove dir", zap.Error(err))
					}
				}
			}
		}()

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
