package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service/out"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
)

type RAUCOfflineService struct {
	SysRoot            string
	InstallRAUCHandler func(raucPath string) error
	CheckSumHandler    out.CheckSumReleaseUseCase

	GetRAUCInfo func(string) (string, error)
}

func (r *RAUCOfflineService) Install(release codegen.Release, sysRoot string) error {
	return r.InstallRAUCHandler(OfflineRAUCFilePath())
}

func (r *RAUCOfflineService) InstallInfo(release codegen.Release, sysRootPath string) (string, error) {
	return OfflineRAUCFilePath(), nil
}

func (r *RAUCOfflineService) LoadReleaseFromRAUC(sysRoot string) (*codegen.Release, error) {
	if _, err := os.Stat(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, common.ReleaseYAMLFileName)); os.IsExist(err) {
		// read release from cache
		return internal.GetReleaseFromLocal(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME))
	}

	rauc_info, err := r.GetRAUCInfo(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME))
	if err != nil {
		return nil, err
	}

	base64_release, err := GetDescription(rauc_info)
	if err != nil {
		return nil, err
	}
	releaseContent, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64_release))
	if err != nil {
		fmt.Println("decode base64 error:", err, "`", strings.TrimSpace(base64_release), "`")
		return nil, err
	}

	release, err := internal.GetReleaseFromContent(releaseContent)
	if err != nil {
		return release, err
	}

	// write release to temp
	err = internal.WriteReleaseToLocal(release, filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME))
	return release, err
}

func (r *RAUCOfflineService) GetRelease(ctx context.Context, tag string, useCache bool) (*codegen.Release, error) {
	if useCache {
		cachePath := filepath.Join(r.SysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME)
		if release, err := internal.GetReleaseFromLocal(cachePath); err == nil {
			return release, nil
		}
	}
	release, err := r.LoadReleaseFromRAUC(r.SysRoot)
	return release, err
}

func (r *RAUCOfflineService) Launch(sysRoot string) error {
	if _, err := os.Stat(filepath.Join(sysRoot, FlagUpgradeFile)); os.IsNotExist(err) {
		return nil
	}

	// remove filepath.Join(sysRoot, FlagUpgradeFile)
	err := os.Remove(filepath.Join(sysRoot, FlagUpgradeFile))

	return err
}

func (r *RAUCOfflineService) VerifyRelease(release codegen.Release) (string, error) {
	return r.CheckSumHandler(release)
}

func CleanupOfflineRAUCTemp(sysRoot string) error {
	return os.RemoveAll(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH))
}

func (r *RAUCOfflineService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	releasePath := filepath.Join(r.SysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
	return releasePath, nil
}

func (r *RAUCOfflineService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	// offline rauc didn't need to extract
	return nil
}

func (r *RAUCOfflineService) PostInstall(release codegen.Release, sysRoot string) error {
	return PostInstallRAUC(release, sysRoot)
}

func (r *RAUCOfflineService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return ShouldUpgrade(release, sysRoot)
}

func (r *RAUCOfflineService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	if !r.ShouldUpgrade(release, sysRootPath) {
		return false
	}

	_, err := r.VerifyRelease(release)
	return err == nil
}

func (r *RAUCOfflineService) PostMigration(sysRoot string) error {
	// return MarkGood()
	// install didn't need to process rauc status now.
	return nil
}

func (r *RAUCOfflineService) Stats() UpdateStats {
	return UpdateStats{
		Name: "Offline RAUC",
	}
}
