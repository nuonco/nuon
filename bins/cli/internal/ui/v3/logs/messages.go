package logs

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second * 5)
	return tickMsg{}
}
