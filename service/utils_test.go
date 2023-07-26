package service_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeVersion(t *testing.T) {
	version := service.NormalizeVersion(common.LegacyWithoutVersion)
	assert.Equal(t, "v0.0.0-legacy-without-version", version)
	_, err := semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("v0.3.5.1")
	assert.Equal(t, "v0.3.5-1", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("v0.3.5.1.1")
	assert.Equal(t, "v0.3.5-1.1", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("v0.3.5")
	assert.Equal(t, "v0.3.5", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("v0.3.5-alpha1")
	assert.Equal(t, "v0.3.5-alpha1", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("v0.3.5-alpha.1")
	assert.Equal(t, "v0.3.5-alpha.1", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("V0.3.5-alpha.1")
	assert.Equal(t, "v0.3.5-alpha.1", version)
	_, err = semver.NewVersion(version)
	assert.NoError(t, err)

	version = service.NormalizeVersion("${DOWNLOAD_DOMAIN}IceWhaleTech/CasaOS/releases/download/v0.3.6/linux-${ARCH}-casaos-migration-tool-v0.3.6.tar.gz")
	_, err = semver.NewVersion(version)
	assert.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func TestVerifyChecksum(t *testing.T) {
	// Create a temp file
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	text := []byte("Hello, World!")
	if _, err := tmpfile.Write(text); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Calculate checksum using sha256sum command
	out, err := exec.Command("sha256sum", tmpfile.Name()).Output() // nolint: gosec
	if err != nil {
		t.Fatal(err)
	}
	checksum := strings.Split(string(out), " ")[0]

	// Test the function
	err = service.VerifyChecksumByFilePath(tmpfile.Name(), checksum)
	assert.NoError(t, err)

	// Test the function with wrong checksum
	err = service.VerifyChecksumByFilePath(tmpfile.Name(), "wrongchecksum")
	assert.Error(t, err)
}
