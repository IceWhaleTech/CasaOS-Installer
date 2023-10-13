package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

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
	releaseContent, err := base64.StdEncoding.DecodeString(base64_release)
	if err != nil {
		return nil, err
	}

	release, err := internal.GetReleaseFromContent(releaseContent)
	if err != nil {
		fmt.Println("write release to temp error:", err)
		return release, nil
	}

	// write release to temp
	err = internal.WriteReleaseToLocal(release, filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, common.ReleaseYAMLFileName))
	return release, err
}

func (r *RAUCOfflineService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	if _, err := os.Stat(filepath.Join(r.SysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)); os.IsExist(err) {
		fmt.Println("rauc file  found")
		return internal.GetReleaseFromLocal(filepath.Join(r.SysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME))
	} else {
		fmt.Println("rauc file not found")

		release, err := r.LoadReleaseFromRAUC(r.SysRoot)
		if err != nil {
			return nil, err
		}
		return release, nil
	}
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
	// 这个offline没有变化
	// return ExtractRAUCRelease(packageFilepath, release)
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

func ExtractOfflineRAUCToTemp(sysRoot string) error {
	// to check temp file exist.
	// TODO should also check rauc file
	if _, err := os.Stat(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH), 0755)
		if err != nil {
			return err
		}

		err = internal.Extract(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME), filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RAUCOfflineService) LoadReleaseFromOfflineRAUC(sysRoot string) (*codegen.Release, error) {
	err := ExtractOfflineRAUCToTemp(sysRoot)
	if err != nil {
		return nil, err
	}

	fmt.Println(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME))
	if _, err := os.Stat(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME)); err != nil {
		return nil, fmt.Errorf("rauc release file not found")
	}

	return internal.GetReleaseFromLocal(filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, config.RAUC_OFFLINE_RELEASE_FILENAME))
}

func (r *RAUCOfflineService) PostMigration(sysRoot string) error {
	// return MarkGood()
	return nil
}
