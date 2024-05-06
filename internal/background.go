package internal

import (
	"context"
	"os"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/hashicorp/go-getter"
	"go.uber.org/zap"
)

func BackgroundPath(version codegen.Version) string {
	return config.BackgroundCachePath + version
}

func DownloadReleaseBackground(url string, version string) {
	// to check if the file exist
	if _, err := os.Stat(BackgroundPath(version)); err == nil {
		return
	}

	// if the background url is nil, return
	// download a url as a file
	getClient := getter.Client{
		Ctx:   context.Background(),
		Dst:   BackgroundPath(version),
		Mode:  getter.ClientModeFile,
		Src:   url,
		Umask: 0o022,
		Options: []getter.ClientOption{
			getter.WithProgress(NewTracker(func(downloaded, totalSize int64) {})),
		},
	}
	err := getClient.Get()
	if err != nil {
		logger.Error("error when trying to download background", zap.Error(err))
	} else {
		logger.Info("download background success", zap.String("url", url), zap.String("version", version))
	}
}
