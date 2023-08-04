package internal_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"gopkg.in/yaml.v3"
)

func TestGetPackageURLByCurrentArch(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	var release codegen.Release

	err := yaml.Unmarshal([]byte(common.SampleReleaseYAML), &release)
	assert.NoError(t, err)

	for _, mirror := range release.Mirrors {
		packageURL, err := internal.GetPackageURLByCurrentArch(release, mirror)
		assert.NoError(t, err)
		assert.NotEmpty(t, packageURL)
	}
}

func TestDownload(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) //

	logger.LogInitConsoleOnly()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "casaos-installer-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	_, err = internal.Download(ctx, tmpDir, "https://github.com/IceWhaleTech/CasaOS-AppStore/releases/download/v0.4.4-alpha10/linux-all-appstore-v0.4.4-alpha10.tar.gz")
	assert.NoError(t, err)
}

func TestDownloadAndExtract(t *testing.T) {
	if _, exists := os.LookupEnv("CI"); exists {
		t.Skip("skipping test in CI environment")
	}

	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	logger.LogInitConsoleOnly()

	packageURL := "https://github.com/IceWhaleTech/get/releases/download/v0.4.4-alpha3/casaos-amd64-v0.4.4-alpha3.tar.gz"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	releaseDir, err := os.MkdirTemp("", "casaos-test-releasedir-*")
	assert.NoError(t, err)
	defer os.RemoveAll(releaseDir)

	packageFilepath, err := internal.Download(ctx, releaseDir, packageURL)
	assert.NoError(t, err)

	err = internal.Extract(packageFilepath, releaseDir)
	assert.NoError(t, err)

	err = internal.BulkExtract(releaseDir)
	assert.NoError(t, err)

	expectedFiles := []string{
		"/usr/bin/casaos",
		"/var/lib/casaos",
	}

	for _, expectedFile := range expectedFiles {
		_, err := os.Stat(filepath.Join(releaseDir, "build", "sysroot", expectedFile))
		assert.NoError(t, err)
	}
}

func TestInstallRelease(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	logger.LogInitConsoleOnly()

	releaseDir, err := os.MkdirTemp("", "casaos-test-releasedir-*")
	assert.NoError(t, err)
	defer os.RemoveAll(releaseDir)

	sysrootPath, err := os.MkdirTemp("", "casaos-test-sysroot-*")
	assert.NoError(t, err)
	defer os.RemoveAll(sysrootPath)

	err = internal.InstallRelease(releaseDir, sysrootPath)
	assert.ErrorIs(t, err, fs.ErrNotExist)

	sourceSysrootPath := filepath.Join(releaseDir, "build", "sysroot")
	err = os.MkdirAll(filepath.Join(sourceSysrootPath, "usr", "bin"), 0o755)
	assert.NoError(t, err)

	err = os.MkdirAll(filepath.Join(sourceSysrootPath, "var", "lib"), 0o755)
	assert.NoError(t, err)

	expectedFiles := []string{
		"/usr/bin/casaos",
		"/var/lib/casaos",
	}

	for _, expectedFile := range expectedFiles {
		err := os.WriteFile(filepath.Join(sourceSysrootPath, expectedFile), []byte{}, 0o600)
		assert.NoError(t, err)
	}

	err = internal.InstallRelease(releaseDir, sysrootPath)
	assert.NoError(t, err)

	for _, expectedFile := range expectedFiles {
		_, err := os.Stat(filepath.Join(sysrootPath, expectedFile))
		assert.NoError(t, err)
	}
}

// NOTE: the test require sudo permission
func TestInstallDocker(t *testing.T) {
	// if environment have non-root permission, skip test
	if os.Geteuid() != 0 {
		t.Skip("skipping test in no-root environment")
	}

	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) //

	logger.LogInitConsoleOnly()

	value := internal.IsDockerInstalled()
	assert.True(t, value)

	value, err := internal.GetDockerRunningStatus()
	assert.NoError(t, err)
	assert.True(t, value)
}
