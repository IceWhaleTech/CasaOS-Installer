package internal_test

import (
	"context"
	"os"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestGetPackageURLByCurrentArch(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	release := codegen.Release{
		Packages: []codegen.Package{
			{
				Architecture: codegen.Amd64,
				URL:          "https://github.com/IceWhaleTech/get/releases/download/v0.4.4-alpha1/casaos-amd64-v0.4.4-alpha1.tar.gz",
			},
			{
				Architecture: codegen.Arm64,
				URL:          "https://github.com/IceWhaleTech/get/releases/download/v0.4.4-alpha1/casaos-arm64-v0.4.4-alpha1.tar.gz",
			},
			{
				Architecture: codegen.Arm7,
				URL:          "https://github.com/IceWhaleTech/get/releases/download/v0.4.4-alpha1/casaos-arm-7-v0.4.4-alpha1.tar.gz",
			},
		},
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release)
	assert.NoError(t, err)
	assert.NotEmpty(t, packageURL)
}

func TestDownloadPackage(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	logger.LogInitConsoleOnly()

	packageURL := "https://github.com/IceWhaleTech/get/releases/download/v0.4.4-alpha1/casaos-amd64-v0.4.4-alpha1.tar.gz"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir, err := internal.DownloadPackage(ctx, packageURL)
	assert.NoError(t, err)

	defer os.RemoveAll(tempDir)
	assert.NotEmpty(t, tempDir)
}
