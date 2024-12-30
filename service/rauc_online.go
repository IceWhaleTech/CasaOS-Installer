package service

import (
	"context"
	"os"
	"path/filepath"
	"reflect"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/checksum"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service/out"
	"go.uber.org/zap"
)

type ChannelType string

const (
	StableChannelType      ChannelType = "stable"
	PublicTestChannelType  ChannelType = "public-test"
	PrivateTestChannelType ChannelType = "private-test"
	TestVerifyChannelType  ChannelType = "test-verification"
	DisableChannelType     ChannelType = "disable"
	UnknownChannelType     ChannelType = "unknown"
)

var (
	ChannelData = map[ChannelType][]string{
		StableChannelType: []string{
			"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/",
			"https://raw.githubusercontent.com/IceWhaleTech/ZimaOS/refs/heads/main/",
		},
		PublicTestChannelType: []string{
			"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/public-test/",
			"https://raw.githubusercontent.com/IceWhaleTech/ZimaOS/refs/heads/main/public-test/",
		},
		PrivateTestChannelType: []string{
			"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/private-test/",
		},
		TestVerifyChannelType: []string{
			"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/test-verification-channel/",
		},
		DisableChannelType: []string{
			"https://localhost/IceWhaleTech/zimaos-rauc/",
		},
	}
)

type RAUCService struct {
	InstallRAUCHandler func(raucPath string) error
	DownloadHandler    out.DownloadReleaseUseCase
	CheckSumHandler    out.CheckSumReleaseUseCase
	URLHandler         ConstructReleaseFileURLFunc
	hasChecked         bool
	path               string
}

func (r *RAUCService) Install(release codegen.Release, sysRoot string) error {
	err := CheckMemory()
	if err != nil {
		return err
	}
	return InstallRAUC(release, sysRoot, r.InstallRAUCHandler)
}

// return the path of update package in local
func (r *RAUCService) InstallInfo(release codegen.Release, sysRootPath string) (string, error) {
	// 	return filepath.Join(config.SysRoot, config.RAUC_RELEASE_PATH, "latest"), nil
	// TODO: fix other case might return err
	return RAUCFilePath(release)
}

func (r *RAUCService) GetRelease(ctx context.Context, tag string, useCache bool) (*codegen.Release, error) {
	if useCache {
		releaseDir, _ := config.ReleaseDir(codegen.Release{Version: "latest"})
		releaseYamlFileName := filepath.Join(releaseDir, common.ReleaseYAMLFileName)
		if release, err := internal.GetReleaseFromLocal(releaseYamlFileName); err == nil {
			logger.Info("get release from local", zap.String("tag", tag), zap.String("path", releaseYamlFileName))
			return release, nil
		}
	}
	return FetchRelease(ctx, tag, r.URLHandler)
}

func (r *RAUCService) VerifyRelease(release codegen.Release) (string, error) {
	_, err := checksum.OnlineRAUCExist(release)
	if err == nil && r.hasChecked {
		return r.path, nil
	}
	path, err := r.CheckSumHandler(release)
	if err == nil {
		r.path = path
		r.hasChecked = true
	}
	return path, err
}

func (r *RAUCService) CleanRelease(ctx context.Context, release codegen.Release) error {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return err
	}
	return os.RemoveAll(releaseDir)
}

func (r *RAUCService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	filePath, err := r.VerifyRelease(release)
	if err != nil {
		logger.Error("verify release error", zap.Error(err))
		// delete the old release
		r.CleanRelease(ctx, release)
	}
	if err == nil {
		return filePath, nil
	}

	_, err = DownloadRelease(ctx, release, force)
	r.hasChecked = false
	if err != nil {
		return "", err
	}
	filePath, err = r.VerifyRelease(release)
	return filePath, err
}

func (r *RAUCService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	// return ExtractRAUCRelease(packageFilepath, release)
	return nil
}

func (r *RAUCService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return ShouldUpgrade(release, sysRoot)
}

func (r *RAUCService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	if !r.ShouldUpgrade(release, sysRootPath) {
		return false
	}

	_, err := r.VerifyRelease(release)
	return err == nil
}

func (r *RAUCService) Launch(sysRoot string) error {
	if _, err := os.Stat(filepath.Join(sysRoot, FlagUpgradeFile)); os.IsNotExist(err) {
		return nil
	}

	// remove filepath.Join(sysRoot, FlagUpgradeFile)
	err := os.Remove(filepath.Join(sysRoot, FlagUpgradeFile))

	return err
}

func (r *RAUCService) PostInstall(release codegen.Release, sysRoot string) error {
	return PostInstallRAUC(release, sysRoot)
}

func (r *RAUCService) PostMigration(sysRoot string) error {
	// return MarkGood()
	return nil
}

func (r *RAUCService) Stats() UpdateServerStats {
	currentChannel := config.ServerInfo.Mirrors
	channelType := UnknownChannelType
	for cType, channel := range ChannelData {
		if reflect.DeepEqual(channel, currentChannel) {
			channelType = cType
		}
	}
	return UpdateServerStats{
		Name:    "Online RAUC",
		Channel: channelType,
	}
}
