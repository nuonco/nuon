package activities

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestImageTagSyncLanded(t *testing.T) {
	assert.True(t, imageTagSyncLanded(app.InstallDeployStatusActive, "bld", "bld"))
	assert.False(t, imageTagSyncLanded(app.InstallDeployStatusError, "bld", "bld"))
	assert.False(t, imageTagSyncLanded(app.InstallDeployStatusCancelled, "bld", "bld"))
	assert.False(t, imageTagSyncLanded(app.InstallDeployStatusActive, "old", "bld"))
	assert.False(t, imageTagSyncLanded(app.InstallDeployStatusActive, "", "bld"))
}
