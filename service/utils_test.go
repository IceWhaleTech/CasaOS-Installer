package service_test

import (
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
}
