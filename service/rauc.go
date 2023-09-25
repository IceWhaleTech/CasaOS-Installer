package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/holoplot/go-rauc/rauc"
)

const (
	FlagUpgradeFile = "/var/lib/casaos/upgradInfo.txt"
)

func ExtractRAUCRelease(packageFilepath string, release codegen.Release) error {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return err
	}

	return internal.BulkExtract(releaseDir)
}

// dependent config.ServerInfo.CachePath
func InstallRAUC(release codegen.Release, sysRoot string, InstallRAUCHandler func(raucPath string) error) error {
	// to check rauc tar

	raucFilePath, err := RAUCFilePath(release)
	if err != nil {
		return err
	}

	err = InstallRAUCHandler(raucFilePath)
	if err != nil {
		log.Fatal("VerifyRAUC() failed: ", err.Error())
	}

	return nil
}

func InstallRAUCImp(raucFilePath string) error {
	// install rauc
	fmt.Println("rauc路径为:", raucFilePath)

	raucInstaller, err := rauc.InstallerNew()
	if err != nil {
		fmt.Sprintln("rauc.InstallerNew() failed: ", err.Error())
	}

	compatible, version, err := raucInstaller.Info(raucFilePath)
	if err != nil {
		log.Fatal("Info() failed", err.Error())
	}
	log.Printf("Info(): compatible=%s, version=%s", compatible, version)

	err = raucInstaller.InstallBundle(raucFilePath, rauc.InstallBundleOptions{})
	if err != nil {
		log.Fatal("InstallBundle() failed: ", err.Error())
	}

	return nil
}

func MockInstallRAUC(raucFilePath string) error {
	// to check file exist
	fmt.Println("文件名为", raucFilePath)
	if _, err := os.Stat(raucFilePath); os.IsNotExist(err) {
		return fmt.Errorf("not found rauc install package")
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

func RAUCFilePath(release codegen.Release) (string, error) {
	// 这个是验证解压之后的包。
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	packageFilePath = packageFilePath[:len(packageFilePath)-len(".tar")] + ".raucb"
	// to check file exist
	fmt.Println("rauc verify in cache:", packageFilePath)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found rauc install package")
	}
	return packageFilePath, nil
}

func MarkGood() error {
	return exec.Command("rauc", "status", "mark-good").Run()
}

func RebootSystem() {
	exec.Command("reboot").Run()
}
