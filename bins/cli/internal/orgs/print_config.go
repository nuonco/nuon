package orgs

import "github.com/powertoolsdev/mono/bins/cli/internal/ui"

func (s *Service) PrintConfig(asJSON bool) {
	view := ui.NewGetView()

	settings := s.cfg.AllSettings()

	if asJSON {
		ui.PrintJSON(settings)
		return
	}

	view.Render([][]string{
		{"Org ID", settings["org_id"].(string)},
		{"App ID", settings["app_id"].(string)},
		{"Install ID", settings["install_id"].(string)},
		{"API Token", settings["api_token"].(string)},
	})
}
