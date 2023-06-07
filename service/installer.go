package service

import (
	"context"
	"fmt"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

const cacheKey = "release"

var (
	Cache              = cache.New(5*time.Minute, 10*time.Minute)
	ErrReleaseNotFound = fmt.Errorf("release not found")
)

func GetRelease(tag string) (*codegen.Release, error) {
	if cached, ok := Cache.Get(cacheKey); ok {
		if release, ok := cached.(*codegen.Release); ok {
			return release, nil
		}
	}

	var release *codegen.Release
	for _, baseURL := range config.ServerInfo.ReleaseBaseURLList {
		releaseURL := fmt.Sprintf("%s/%s/casaos-release", baseURL, tag)

		logger.Info("trying to get release information from url", zap.String("url", releaseURL))

		_release, err := internal.GetReleaseFrom(releaseURL)
		if err != nil {
			logger.Info("error while getting release information - skipping", zap.Error(err), zap.String("url", releaseURL))
			continue
		}

		release = _release
		break
	}

	if release == nil {
		return nil, ErrReleaseNotFound
	}

	Cache.Set(cacheKey, release, cache.DefaultExpiration)

	return release, nil
}

func InstallRelease(ctx echo.Context, release codegen.Release) error {
	// TODO: get releaseDir based on release

	// TODO: if releaseDir does not exist, download and extract package to releaseDir
	releaseDir := "TODO"

	// TODO: write release information to releaseDir/release.yaml

	backgroundCtx := context.Background()
	if err := internal.InstallRelease(backgroundCtx, releaseDir, "/"); err != nil {
		return err
	}

	return nil
}
