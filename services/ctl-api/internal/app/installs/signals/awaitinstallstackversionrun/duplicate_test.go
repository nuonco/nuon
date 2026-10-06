package awaitinstallstackversionrun

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestIsDuplicateOfActive(t *testing.T) {
	status := func(s app.Status, md map[string]any) *app.InstallStackVersion {
		return &app.InstallStackVersion{Status: app.CompositeStatus{Status: s, Metadata: md}}
	}
	assert.True(t, isDuplicateOfActive(status(app.InstallStackVersionStatusOutdated, map[string]any{app.InstallStackVersionDuplicateOfMetadataKey: "istactive"})))
	assert.False(t, isDuplicateOfActive(status(app.InstallStackVersionStatusOutdated, nil)))
	assert.False(t, isDuplicateOfActive(status(app.InstallStackVersionStatusPendingUser, map[string]any{app.InstallStackVersionDuplicateOfMetadataKey: "istactive"})))
}
