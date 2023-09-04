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

const (
	RAUC_OFFLINE_PATH             = "/DATA/rauc/"
	RAUC_OFFLINE_RELEASE_FILENAME = "release.yaml"
	RAUC_OFFLINE_RAUC_FILENAME    = "rauc.tar.gz"
	OFFLINE_RAUC_TEMP_PATH        = "/tmp/offline_rauc"
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
	fmt.Println(filepath.Join(r.SysRoot, RAUC_OFFLINE_PATH, RAUC_OFFLINE_RAUC_FILENAME))

	if _, err := os.Stat(filepath.Join(r.SysRoot, RAUC_OFFLINE_PATH, RAUC_OFFLINE_RAUC_FILENAME)); os.IsNotExist(err) {
		return nil, fmt.Errorf("not found offline install package")
	} else {
		fmt.Println("found offline install package")
	}

	release, err := r.LoadReleaseFromOfflineRAUC(r.SysRoot)
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
	return VerifyRAUCOfflineRelease(release)
}
func VerifyRAUCOfflineRelease(release codegen.Release) (string, error) {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageFilePath := filepath.Join(releaseDir, RAUC_OFFLINE_RAUC_FILENAME)

	// to check file exist
	fmt.Println("rauc verify:", packageFilePath)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found offline rauc release package")
	}
	return packageFilePath, nil

}

func (r *RAUCOfflineService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	releasepath, err := r.VerifyRelease(release)
	if err != nil {
		// 这里多做一步，从本地读release
		// 把前面的zip复制到/var/cache/casaos下面。
		releaseDir, err := ReleaseDir(release)
		if err != nil {
			return "", err
		}
		//copy file to /var/cache/casaos
		os.MkdirAll(releaseDir, 0755)
		_, err = copy(filepath.Join(r.SysRoot, RAUC_OFFLINE_PATH, RAUC_OFFLINE_RAUC_FILENAME), filepath.Join(releaseDir, RAUC_OFFLINE_RAUC_FILENAME))
		if err != nil {
			return "", err
		}

		return filepath.Join(releaseDir, RAUC_OFFLINE_RAUC_FILENAME), nil
	}
	return releasepath, nil
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

	_, err := r.VerifyRelease(release)
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

func ExtractOfflineRAUCToTemp(sysRoot string) error {
	// to check temp file exist.
	// TODO should also check rauc file
	if _, err := os.Stat(filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH, RAUC_OFFLINE_RELEASE_FILENAME)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH), 0755)
		if err != nil {
			return err
		}

		err = internal.Extract(filepath.Join(sysRoot, RAUC_OFFLINE_PATH, RAUC_OFFLINE_RAUC_FILENAME), filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH))
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

	fmt.Println(filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH, RAUC_OFFLINE_RELEASE_FILENAME))
	if _, err := os.Stat(filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH, RAUC_OFFLINE_RELEASE_FILENAME)); err != nil {
		return nil, fmt.Errorf("rauc release file not found")
	}

	release, err := internal.GetReleaseFromLocal(filepath.Join(sysRoot, OFFLINE_RAUC_TEMP_PATH, RAUC_OFFLINE_RELEASE_FILENAME))
	if err != nil {
		return nil, err
	}
	return release, nil
}
