package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, appID string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		res, err := s.api.DeleteApp(ctx, appID)
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: appID, Deleted: res}
		ui.PrintJSON(r)
		return
	}

	view := ui.NewDeleteView("app", appID)
	view.Start()
	view.Update("deleting app")

	_, err = s.api.DeleteApp(ctx, appID)
	if err != nil {
		view.Fail(err)
		return
	}
	view.SuccessQueued()
}
