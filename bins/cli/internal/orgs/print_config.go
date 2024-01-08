package orgs

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) PrintConfig(asJSON bool) {
	view := ui.NewGetView()

	settings := s.cfg.AllSettings()
	if len(settings) == 0 {
		pterm.DefaultBasicText.Println("No config set")
		return
	} else {

		if asJSON {
			ui.PrintJSON(settings)
			return
		}

		var data = [][]string{}
		for k, v := range settings {
			data = append(data, []string{k, v.(string)})
		}

		view.Render(data)
	}
}
