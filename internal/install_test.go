package internal_test

import (
	"testing"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/stretchr/testify/assert"
)

func TestGetPackageURLByCurrentArch(t *testing.T) {
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
