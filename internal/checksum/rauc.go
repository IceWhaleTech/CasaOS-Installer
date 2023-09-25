package checksum

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

func OnlineTarExist(release codegen.Release) (string, error) {
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

	// to check file exist
	fmt.Println("rauc verify release:", packageFilePath)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found rauc release  package")
	}
	return packageFilePath, nil
}

func OfflineTarExist(release codegen.Release) (string, error) {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageFilePath := filepath.Join(releaseDir, config.RAUC_OFFLINE_RAUC_FILENAME)

	// to check file exist
	fmt.Println("rauc  offline verify in cache:", packageFilePath)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found offline rauc release package")
	}
	return packageFilePath, nil

}
