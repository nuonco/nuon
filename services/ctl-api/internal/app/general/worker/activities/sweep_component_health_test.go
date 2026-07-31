package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The sweep only marks installs unknown; recovery arrives by push. So the
// window has to exclude two things: installs merely between reports, and dead
// installs already marked unknown, which would otherwise be re-swept forever.
func TestComponentHealthSweepWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	quietBefore, ignoreBefore := componentHealthSweepWindow(now)

	reportedAt := func(ago time.Duration) bool {
		at := now.Add(-ago)
		return at.Before(quietBefore) && at.After(ignoreBefore)
	}

	assert.False(t, reportedAt(1*time.Minute), "between reports, not stale")
	assert.False(t, reportedAt(4*time.Minute), "still inside the staleness threshold")
	assert.True(t, reportedAt(6*time.Minute), "quiet past the threshold — sweep it")
	assert.True(t, reportedAt(3*time.Hour), "still inside the band")
	assert.False(t, reportedAt(7*time.Hour), "already unknown; re-sweeping changes nothing")

	assert.Equal(t, componentHealthStaleAfter, now.Sub(quietBefore),
		"sweep threshold must match the evaluator's staleness threshold")
}
