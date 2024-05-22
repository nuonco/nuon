package components

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, appID, compID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		res, err := s.api.DeleteComponent(ctx, compID)
		if err != nil {
			ui.PrintJSONError(err)
			return
		}

		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: compID, Deleted: res}
		ui.PrintJSON(r)
		return
	}

	view := ui.NewDeleteView("component", compID)
	view.Start()

	_, err = s.api.DeleteComponent(ctx, compID)
	if err != nil {
		view.Fail(err)
		return
	}
	view.Success()
}
