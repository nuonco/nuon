package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) CurrentInputs(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}
	view := ui.NewGetView()

	inputs, err := s.api.GetInstallInputs(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(inputs)
		return
	}

	for _, inp := range inputs {
		data := [][]string{}
		for k, v := range inp.Values {
			data = append(data, []string{k, v})
		}
		pterm.Println("")
		pterm.DefaultBasicText.Println("inputs ID: " + pterm.LightMagenta(inp.ID))
		view.Render(data)
	}
}
