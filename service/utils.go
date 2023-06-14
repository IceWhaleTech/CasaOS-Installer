package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/Masterminds/semver/v3"
)

func MigrationToolsDir() string {
	return filepath.Join(config.ServerInfo.CachePath, "migration-tools")
}

func ReleaseDir(release codegen.Release) (string, error) {
	if release.Version == "" {
		return "", fmt.Errorf("release version is empty")
	}

	return filepath.Join(config.ServerInfo.CachePath, "releases", release.Version), nil
}

func NormalizeVersion(version string) string {
	if version == common.LegacyWithoutVersion {
		return "v0.0.0-legacy-without-version"
	}

	version = "v" + strings.TrimLeft(version, "Vv")

	if _, err := semver.NewVersion(version); err == nil {
		return version
	}

	versionNumbers := strings.SplitN(version, ".", 3)
	if len(versionNumbers) < 3 {
		return version
	}

	versionNumbers[2] = strings.Replace(versionNumbers[2], ".", "-", 1)

	return strings.Join([]string{versionNumbers[0], versionNumbers[1], versionNumbers[2]}, ".")
}

// sha256sum
func VerifyChecksum(filepath, checksum string) error {
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
