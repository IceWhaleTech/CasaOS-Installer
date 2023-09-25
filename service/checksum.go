package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

func DownloadChecksum(ctx context.Context, release codegen.Release, mirror string) (string, error) {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return "", err
	}

	checksumURL := internal.GetChecksumsURL(release, mirror)
	return internal.Download(ctx, releaseDir, checksumURL)
}

// sha256sum
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

func GetChecksums(release codegen.Release) (map[string]string, error) {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return nil, err
	}

	checksumsFilePath := filepath.Join(releaseDir, common.ChecksumsTXTFileName)

	return internal.GetChecksums(checksumsFilePath)
}
