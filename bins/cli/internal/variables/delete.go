package variables

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, appID, variableID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		res, err := s.api.DeleteAppSecret(ctx, appID, variableID)
		if err != nil {
			return ui.PrintJSONError(err)
		}

		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: variableID, Deleted: res}
		ui.PrintJSON(r)
		return nil
	}

	view := ui.NewDeleteView("variable", variableID)
	view.Start()
	_, err = s.api.DeleteAppSecret(ctx, appID, variableID)
	if err != nil {
		return view.Fail(err)
	}
	view.Success()
	return nil
}
