package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/holoplot/go-rauc/rauc"
)

const (
	RAUCOfflinePath        = "/Data/rauc/"
	RAUCOfflineReleaseFile = "release.yaml"
	RAUCOfflineRAUCFile    = "rauc.tar.gz"

	FlagUpgradeFile = "/var/lib/casaos/upgradInfo.txt"
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
	return VerifyRAUC(release)
}

func (r *RAUCService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	// TODO 这里重做，不做额外的功能
	return DownloadRelease(ctx, release, force)
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

func ExtractRAUCRelease(packageFilepath string, release codegen.Release) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	return internal.BulkExtract(releaseDir)
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

func LoadReleaseFromLocal(sysRoot string) (*codegen.Release, error) {
	// to check RAUCOfflinePath + RAUCOfflineReleaseFile
	fmt.Println(filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineReleaseFile))
	if _, err := os.Stat(filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineReleaseFile)); err != nil {
		return nil, fmt.Errorf("rauc release file not found")
	}

	release, err := internal.GetReleaseFromLocal(filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineReleaseFile))
	if err != nil {
		return nil, err
	}
	return release, nil
}

// dependent config.ServerInfo.CachePath
func InstallRAUC(release codegen.Release, sysRoot string, InstallRAUCHandler func(raucPath string) error) error {
	// to check rauc tar

	raucFilePath, err := VerifyRAUC(release)
	if err != nil {
		return err
	}

	err = InstallRAUCHandler(raucFilePath)
	if err != nil {
		log.Fatal("VerifyRAUC() failed: ", err.Error())
	}

	return nil
}

func InstallRAUCHandlerV1(RAUCFilePath string) error {
	// install rauc
	fmt.Println("rauc路径为:", RAUCFilePath)

	raucInstaller, err := rauc.InstallerNew()
	if err != nil {
		fmt.Sprintln("rauc.InstallerNew() failed: ", err.Error())
	}

	compatible, version, err := raucInstaller.Info(RAUCFilePath)
	if err != nil {
		log.Fatal("Info() failed", err.Error())
	}
	log.Printf("Info(): compatible=%s, version=%s", compatible, version)

	err = raucInstaller.InstallBundle(RAUCFilePath, rauc.InstallBundleOptions{})
	if err != nil {
		log.Fatal("InstallBundle() failed: ", err.Error())
	}

	return nil
}

func InstallRAUCTest(raucfilepath string) error {
	// to check file exist
	fmt.Println("文件名为", raucfilepath)
	if _, err := os.Stat(raucfilepath); os.IsNotExist(err) {
		return fmt.Errorf("not found offline install package")
	}

	return nil
}

func PostInstallRAUC(release codegen.Release, sysRoot string) error {
	// write 1+1=2  to sysRoot + FlagUpgradeFile
	d1 := []byte("1+1=2")
	err := os.WriteFile(filepath.Join(sysRoot, FlagUpgradeFile), d1, 0644)

	RebootSystem()
	return err
}

func VerifyRAUC(release codegen.Release) (string, error) {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	// 不能判断tar.gz在不在，因为离线包的名字不一样
	// if _, err := os.Stat(packageFilePath); err != nil {
	// 	return "", fmt.Errorf("rauc %s not found", packageFilePath)
	// }

	// 这里需要注意raucb的名字必须和包名一致
	// TODO 更好的包信息，不能只有包名，没有rauc名。
	// replace tar.gz to raucb of packageFilePath
	packageFilePath = packageFilePath[:len(packageFilePath)-len(".tar.gz")] + ".raucb"
	return packageFilePath, nil
}

func RebootSystem() {
	exec.Command("reboot").Run()
}
