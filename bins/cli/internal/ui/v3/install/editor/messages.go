package editor

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type configFetchedMsg struct {
	inputConfig   *models.AppAppInputConfig
	install       *models.AppInstall
	currentInputs *models.AppInstallInputs
	err           error
}

type inputsUpdatedMsg struct {
	inputs *models.AppInstallInputs
	err    error
}

type autoExitMsg struct{}

func autoExitAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return autoExitMsg{}
	})
}
