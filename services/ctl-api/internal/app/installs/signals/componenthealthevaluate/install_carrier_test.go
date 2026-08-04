package componenthealthevaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func TestInstallCarrierIndex(t *testing.T) {
	resp := func(installHealth string, recovered ...bool) *activities.EvaluateComponentHealthResponse {
		r := &activities.EvaluateComponentHealthResponse{}
		if installHealth != "" {
			r.InstallNotification = &activities.InstallHealthNotification{Health: installHealth}
		}
		for _, rec := range recovered {
			r.Notifications = append(r.Notifications, activities.ComponentHealthNotification{Recovered: rec})
		}
		return r
	}

	t.Run("no install crossing", func(t *testing.T) {
		assert.Equal(t, -1, installCarrierIndex(resp("", false)))
	})

	t.Run("degrade rides the failing component", func(t *testing.T) {
		assert.Equal(t, 1, installCarrierIndex(resp("degraded", true, false)))
	})

	t.Run("recovery rides the recovering component", func(t *testing.T) {
		assert.Equal(t, 1, installCarrierIndex(resp("healthy", false, true)))
	})

	// A crossing caused by a component whose alert was suppressed has nothing to
	// ride on, so the standalone install alert must still fire.
	t.Run("no matching component keeps the standalone alert", func(t *testing.T) {
		assert.Equal(t, -1, installCarrierIndex(resp("degraded", true)))
		assert.Equal(t, -1, installCarrierIndex(resp("degraded")))
	})
}
