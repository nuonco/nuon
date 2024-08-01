package apps

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		res, err := s.api.DeleteApp(ctx, appID)
		if err != nil {
			ui.PrintJSONError(err)
			return err
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: appID, Deleted: res}
		ui.PrintJSON(r)
		return nil
	}

	view := ui.NewDeleteView("app", appID)
	view.Start()
	view.Update("deleting app")

	_, err = s.api.DeleteApp(ctx, appID)
	if err != nil {
		return view.Fail(err)
	}
	view.SuccessQueued()
	return nil
}
