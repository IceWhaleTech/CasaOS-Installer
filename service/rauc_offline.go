package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
)

type RAUCOfflineService struct {
	SysRoot            string
	InstallRAUCHandler func(raucPath string) error
}

func (r *RAUCOfflineService) Install(release codegen.Release, sysRoot string) error {
	return InstallRAUC(release, sysRoot, r.InstallRAUCHandler)
}

func (r *RAUCOfflineService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {

	// to check file exist
	fmt.Println(filepath.Join(r.SysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile))

	if _, err := os.Stat(filepath.Join(r.SysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile)); os.IsNotExist(err) {
		return nil, fmt.Errorf("not found offline install package")
	} else {
		fmt.Println("found offline install package")
	}

	err := internal.Extract(filepath.Join(r.SysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile), filepath.Join(r.SysRoot, RAUCOfflinePath))
	if err != nil {
		return nil, err
	}

	// TODO 这里改成去解压拿release
	release, err := LoadReleaseFromLocal(r.SysRoot)
	if err != nil {
		return nil, err
	}
	return release, nil
}
func (r *RAUCOfflineService) MigrationInLaunch(sysRoot string) error {
	if _, err := os.Stat(filepath.Join(sysRoot, FlagUpgradeFile)); os.IsNotExist(err) {
		return nil
	}

	// remove filepath.Join(sysRoot, FlagUpgradeFile)
	err := os.Remove(filepath.Join(sysRoot, FlagUpgradeFile))

	return err
}

func (r *RAUCOfflineService) VerifyRelease(release codegen.Release) (string, error) {
	return VerifyRAUC(release)
}

func (r *RAUCOfflineService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	// 这里多做一步，从本地读release
	// 把前面的zip复制到/var/cache/casaos下面。
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}
	//copy file to /var/cache/casaos
	os.MkdirAll(releaseDir, 0755)
	_, err = copy(filepath.Join(r.SysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile), filepath.Join(releaseDir, RAUCOfflineRAUCFile))
	if err != nil {
		return "", err
	}

	return filepath.Join(releaseDir, RAUCOfflineRAUCFile), nil
}

func (r *RAUCOfflineService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	// 这个offline没有变化
	return ExtractRAUCRelease(packageFilepath, release)
}

func (r *RAUCOfflineService) PostInstall(release codegen.Release, sysRoot string) error {
	return PostInstallRAUC(release, sysRoot)
}

func (r *RAUCOfflineService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return ShouldUpgrade(release, sysRoot)
}
func (r *RAUCOfflineService) IsUpgradable(release codegen.Release, sysrootPath string) bool {

	if !r.ShouldUpgrade(release, sysrootPath) {
		return false
	}

	_, err := VerifyRAUC(release)
	return err == nil
}

func (r *RAUCOfflineService) GetMigrationInfo(ctx context.Context, release codegen.Release) error {
	return nil
}

func (r *RAUCOfflineService) DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
