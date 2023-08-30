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

// 目前不需要这个做额外的动作
func DownloadReleaseZIP(release codegen.Release, force bool) (string, error) {
	return "", nil
}

func ExtractRAUCRelease(packageFilepath string, release codegen.Release) error {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return err
	}

	if err := internal.Extract(packageFilepath, releaseDir); err != nil {
		return err
	}

	return internal.BulkExtract(releaseDir)
}

func (r *RAUCService) MigrationInLaunch(sysRoot string) error {
	return StartMigration(sysRoot)
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

	RebootSystem()
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

	if _, err := os.Stat(packageFilePath); err != nil {
		return "", fmt.Errorf("rauc %s not found", packageFilePath)
	}

	// TODO 更好的包信息，不能只有包名，没有rauc名。
	// replace tar.gz to raucb of packageFilePath
	packageFilePath = packageFilePath[:len(packageFilePath)-len(".tar.gz")] + ".raucb"
	return packageFilePath, nil
}

func RebootSystem() {
	exec.Command("reboot").Run()
}
