package common

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const DefaultRefreshInterval = 5 * time.Second

type TickMsg time.Time

func TickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
