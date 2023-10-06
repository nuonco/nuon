package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, id string, asJSON bool) {
	if asJSON {
		res, err := s.api.DeleteInstall(ctx, id)
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: id, Deleted: res}
		ui.PrintJSON(r)
		return
	}

	view := ui.NewDeleteView("install", id)
	view.Start()
	_, err := s.api.DeleteInstall(ctx, id)
	if err != nil {
		view.Fail(err)
		return
	}
	view.Success()
}
