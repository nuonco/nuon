package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPartitionStaleSignals(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	ptr := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name    string
		signals []*app.QueueSignal
		maxAge  time.Duration
		stale   map[string][]string
		live    []string
	}{
		{
			name:    "fresh signal with no expiry holds the emitter",
			signals: []*app.QueueSignal{{ID: "a", CreatedAt: now.Add(-time.Minute)}},
			live:    []string{"a"},
			stale:   map[string][]string{},
		},
		{
			name:    "unexpired signal holds the emitter",
			signals: []*app.QueueSignal{{ID: "a", CreatedAt: now.Add(-time.Minute), ExpiresAt: ptr(now.Add(time.Minute))}},
			live:    []string{"a"},
			stale:   map[string][]string{},
		},
		// The wedge: a dead handler never enforces its own expiry, so without this the
		// emitter is blocked forever and the cron silently stops.
		{
			name:    "expired signal releases the emitter",
			signals: []*app.QueueSignal{{ID: "a", CreatedAt: now.Add(-time.Hour), ExpiresAt: ptr(now.Add(-time.Minute))}},
			live:    []string{},
			stale:   map[string][]string{staleReasonExpired: {"a"}},
		},
		{
			name:    "max-in-flight-age still applies without an expiry",
			signals: []*app.QueueSignal{{ID: "a", CreatedAt: now.Add(-time.Hour)}},
			maxAge:  30 * time.Minute,
			live:    []string{},
			stale:   map[string][]string{staleReasonMaxInFlightAge: {"a"}},
		},
		{
			name: "each stale signal is reported under its own reason",
			signals: []*app.QueueSignal{
				{ID: "old", CreatedAt: now.Add(-time.Hour)},
				{ID: "gone", CreatedAt: now.Add(-time.Minute), ExpiresAt: ptr(now.Add(-time.Second))},
				{ID: "fine", CreatedAt: now.Add(-time.Minute), ExpiresAt: ptr(now.Add(time.Hour))},
			},
			maxAge: 30 * time.Minute,
			live:   []string{"fine"},
			stale: map[string][]string{
				staleReasonMaxInFlightAge: {"old"},
				staleReasonExpired:        {"gone"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stale, live := partitionStaleSignals(tt.signals, tt.maxAge, now)

			liveIDs := make([]string, 0, len(live))
			for _, s := range live {
				liveIDs = append(liveIDs, s.ID)
			}

			assert.Equal(t, tt.stale, stale)
			assert.Equal(t, tt.live, liveIDs)
		})
	}
}
