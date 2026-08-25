package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestInstallRegistryTagIsStableAcrossDeploysOfSameBuild(t *testing.T) {
	first := &app.InstallDeploy{ID: "dpl1", ComponentBuildID: "bld1"}
	second := &app.InstallDeploy{ID: "dpl2", ComponentBuildID: "bld1"}

	assert.Equal(t, installRegistryTag(first), installRegistryTag(second))
	assert.NotEqual(t, first.ID, installRegistryTag(first))
}

func TestInstallRegistryTagDistinguishesBuilds(t *testing.T) {
	old := &app.InstallDeploy{ID: "dpl1", ComponentBuildID: "bld1"}
	rebuilt := &app.InstallDeploy{ID: "dpl2", ComponentBuildID: "bld2"}

	assert.NotEqual(t, installRegistryTag(old), installRegistryTag(rebuilt))
}
