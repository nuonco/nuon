package common

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// DefaultRefreshInterval is the standard polling interval for all TUI data fetching.
const DefaultRefreshInterval = 5 * time.Second

// TickMsg is the shared message type sent on each polling tick.
type TickMsg time.Time

// TickCmd returns a tea.Cmd that sends a TickMsg after the given duration.
func TickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
