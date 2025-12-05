package orgs

import (
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) PrintConfig(asJSON bool) error {
	view := ui.NewGetView()

	settings := s.cfg.AllSettings()
	if len(settings) == 0 {
		fmt.Println("No config set")
		return nil
	} else {

		if asJSON {
			ui.PrintJSON(settings)
			return nil
		}

		var data = [][]string{}
		for k, v := range settings {
			data = append(data, []string{k, v.(string)})
		}

		view.Render(data)
	}
	return nil
}
