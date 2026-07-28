package orgs

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) PrintConfig(asJSON bool) error {
	settings := s.cfg.AllSettings()

	if asJSON {
		ui.PrintJSON(settings)
		return nil
	}

	if len(settings) == 0 {
		ui.Println("No config set")
		return nil
	}

	var data = [][]string{}
	for k, v := range settings {
		data = append(data, []string{k, v.(string)})
	}

	ui.NewGetView().Render(data)
	return nil
}
