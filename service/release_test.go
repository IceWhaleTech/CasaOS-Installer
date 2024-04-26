package service_test

import (
	"testing"

	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestReleaseCompare(t *testing.T) {
	target, err := semver.NewVersion(service.NormalizeVersion("0.4.8"))
	assert.NoError(t, err)
	assert.Equal(t, "0.4.8", target.String())

	current, err := semver.NewVersion(service.NormalizeVersion("0.4.8-5"))
	assert.NoError(t, err)
	assert.Equal(t, "0.4.8-5", current.String())

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.8"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.8"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))
	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.8"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.8"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-1"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-1"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.10-alpha2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.10-alpha2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-alpha1"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-alpha2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-alpha5"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-beta2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-beta2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-alpha5"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-alpha5"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-2"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-beta5"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	//
	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-beta5"))
	assert.Equal(t, true, service.IsNewerVersion(current, target))

	target, _ = semver.NewVersion(service.NormalizeVersion("0.4.9-beta5"))
	current, _ = semver.NewVersion(service.NormalizeVersion("0.4.9"))
	assert.Equal(t, false, service.IsNewerVersion(current, target))
}
