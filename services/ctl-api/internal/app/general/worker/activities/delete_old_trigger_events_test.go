package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestTriggerEventRetention(t *testing.T) {
	tests := map[app.EventRoutingStatus]time.Duration{
		app.EventRoutingStatusRejected: 7 * 24 * time.Hour,
		app.EventRoutingStatusIgnored:  30 * 24 * time.Hour,
	}
	for status, expected := range tests {
		retention, err := triggerEventRetention(status)
		require.NoError(t, err)
		require.Equal(t, expected, retention)
	}

	_, err := triggerEventRetention(app.EventRoutingStatusMatched)
	require.ErrorContains(t, err, "unsupported trigger event cleanup status")
}
