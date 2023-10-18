package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

func OnlineRAUCExist(release codegen.Release) (string, error) {
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

func VerifyChecksumByFilePath(filepath, checksum string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	buf := hash.Sum(nil)[:32]
	if hex.EncodeToString(buf) != checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", checksum, hex.EncodeToString(buf))
	}

	return nil
}

func OnlineRaucChecksumExist(release codegen.Release) (string, error) {
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
	checksums, err := internal.GetChecksums(packageFilePath)
	packageChecksum := checksums[packageFilename]

	if err != nil {
		return "", err
	}
	// to check file exist
	fmt.Println("rauc verify release:", packageFilePath)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found rauc release  package")
	}

	return packageFilePath, VerifyChecksumByFilePath(packageFilePath, packageChecksum)
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

func OfflineTarExistV2(release codegen.Release) (string, error) {
	packageFilePath := filepath.Join(config.SysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found offline rauc release package")
	}
	return packageFilePath, nil
}
