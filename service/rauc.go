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
	RAUCOfflineReleaseFile = "rauc-release.yml"
	RAUCOfflineRAUCFile    = "rauc.tar.gz"
)

type RAUCService struct {
}

func (r *RAUCService) Install(release codegen.Release, sysRoot string) error {
	return InstallRAUC(release, sysRoot)
}

func (r *RAUCService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	// 这里多做一步，从本地读release
	release, err := LoadReleaseFromLocal()
	if err != nil {
		// 不然就是从网络读
		return GetRelease(ctx, tag)
	}
	return release, nil
}
func (r *RAUCService) VerifyRelease(release codegen.Release) (string, error) {
	return VerifyRAUC(release)
}

func (r *RAUCService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	// 这里多做一步，从本地读release
	return DownloadRelease(ctx, release, force)
}

func (r *RAUCService) MigrationInLaunch(sysRoot string) error {
	return StartMigration(sysRoot)
}

func LoadReleaseFromLocal() (*codegen.Release, error) {
	// to check RAUCOfflinePath + RAUCOfflineReleaseFile
	if _, err := os.Stat(RAUCOfflinePath + RAUCOfflineReleaseFile); err != nil {
		return nil, fmt.Errorf("rauc release file not found")
	}

	if _, err := os.Stat(RAUCOfflinePath + RAUCOfflineRAUCFile); err != nil {
		return nil, fmt.Errorf("rauc tar file not found")
	}

	release, err := internal.GetReleaseFromLocal(filepath.Join(RAUCOfflinePath, RAUCOfflineReleaseFile))
	if err != nil {
		return nil, err
	}
	return release, nil
}

// dependent config.ServerInfo.CachePath
func InstallRAUC(release codegen.Release, sysRoot string) error {
	// to check rauc tar

	raucfilepath, err := VerifyRAUC(release)
	if err != nil {
		log.Fatal("VerifyRAUC() failed: ", err.Error())
	}

	// install rauc
	raucInstaller, err := rauc.InstallerNew()
	if err != nil {
		fmt.Sprintln("rauc.InstallerNew() failed: ", err.Error())
	}

	compatible, version, err := raucInstaller.Info(raucfilepath)
	if err != nil {
		log.Fatal("Info() failed", err.Error())
	}
	log.Printf("Info(): compatible=%s, version=%s", compatible, version)

	err = raucInstaller.InstallBundle(raucfilepath, rauc.InstallBundleOptions{})
	if err != nil {
		log.Fatal("InstallBundle() failed: ", err.Error())
	}

	return nil
}

func VerifyRAUC(release codegen.Release) (string, error) {
	releaseDir, err := ReleaseDir(release)
	if err != nil {
		return "", err
	}

	// packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	// if err != nil {
	// 	return "", err
	// }

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	if _, err := os.Stat(packageFilePath); err != nil {
		return "", fmt.Errorf("rauc %s not found", packageFilePath)
	}

	return packageFilePath, nil
}

func RebootSystem() {
	exec.Command("reboot").Run()
}
