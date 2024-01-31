package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if asJSON {
		res, err := s.api.DeleteInstall(ctx, installID)
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: installID, Deleted: res}
		ui.PrintJSON(r)
		return
	}

	view := ui.NewDeleteView("install", installID)
	view.Start()
	_, err = s.api.DeleteInstall(ctx, installID)
	if err != nil {
		view.Fail(err)
		return
	}
	view.SuccessQueued()
}
