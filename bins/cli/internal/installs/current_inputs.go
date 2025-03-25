package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
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

	inputs, err := s.listInstallInputs(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(inputs)
		return nil
	}

	for _, inp := range inputs {
		data := [][]string{}
		for k, v := range inp.RedactedValues {
			data = append(data, []string{k, v})
		}
		pterm.Println("")
		pterm.DefaultBasicText.Println("inputs ID: " + pterm.LightMagenta(inp.ID))
		view.Render(data)
	}
	return nil
}

func (s *Service) listInstallInputs(ctx context.Context, installID string) ([]*models.AppInstallInputs, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetInstallInputs(ctx, installID, &models.GetInstallInputsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstallInputs, bool, error) {
		cmps, hasMore, err := s.api.GetInstallInputs(ctx, installID, &models.GetInstallInputsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return cmps, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
