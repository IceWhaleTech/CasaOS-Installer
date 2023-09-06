package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type RAUCService struct {
	InstallRAUCHandler func(raucPath string) error
}

func (r *RAUCService) Install(release codegen.Release, sysRoot string) error {
	return InstallRAUC(release, sysRoot, r.InstallRAUCHandler)
}

func (r *RAUCService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	return GetRelease(ctx, tag)
}

func (r *RAUCService) VerifyRelease(release codegen.Release) (string, error) {
	// 这个是验证下载包的，验证的是下载之前的包。
	return VerifyRAUCRelease(release)
}

func (r *RAUCService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	filepath, err := r.VerifyRelease(release)
	if err != nil {
		fmt.Println("重新下载")
		return DownloadRelease(ctx, release, force)
	} else {
		fmt.Println("不用下载")
	}

	return filepath, nil
}

func (r *RAUCService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	return ExtractRAUCRelease(packageFilepath, release)
}

func (r *RAUCService) GetMigrationInfo(ctx context.Context, release codegen.Release) error {
	return nil
}

func (r *RAUCService) DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	return nil
}

func (r *RAUCService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return ShouldUpgrade(release, sysRoot)
}

func (r *RAUCService) IsUpgradable(release codegen.Release, sysrootPath string) bool {
	if !r.ShouldUpgrade(release, sysrootPath) {
		return false
	}

	_, err := r.VerifyRelease(release)
	return err == nil
}

func (r *RAUCService) MigrationInLaunch(sysRoot string) error {
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
	return nil
}
