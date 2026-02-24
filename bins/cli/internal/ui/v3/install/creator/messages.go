package creator

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type configFetchedMsg struct {
	inputConfig *models.AppAppInputConfig
	app         *models.AppApp
	err         error
}

type installCreatedMsg struct {
	install *models.AppInstall
	err     error
}

type autoExitMsg struct{}

func autoExitAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return autoExitMsg{}
	})
}
