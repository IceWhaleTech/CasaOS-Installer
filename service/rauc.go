package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/holoplot/go-rauc/rauc"
)

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

	// packageFilename := filepath.Base(packageURL)
	packageFilename := "casaos_generic-x86-64-0.4.4.raucb"

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	if _, err := os.Stat(packageFilePath); err != nil {
		return "", fmt.Errorf("rauc %s not found", packageFilePath)
	}

	return packageFilePath, nil
}

func RebootSystem() {
	exec.Command("reboot").Run()
}
