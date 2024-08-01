package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) CurrentInputs(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewGetView()

	inputs, err := s.api.GetInstallInputs(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(inputs)
		return nil
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
	return nil
}
