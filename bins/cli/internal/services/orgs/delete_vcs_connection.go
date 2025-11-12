package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeleteVCSConnection(ctx context.Context, connID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	connID, err := lookup.VCSConnectionID(ctx, s.api, connID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		err = s.api.DeleteVCSConnection(ctx, connID)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		ui.PrintJSON(response{
			ID:      connID,
			Deleted: true,
		})
		return nil
	} else {
		view := ui.NewDeleteView("github-connection", connID)
		view.Start()
		err := s.api.DeleteVCSConnection(ctx, connID)
		if err != nil {
			return view.Fail(err)
		}

		view.Success()
		return nil
	}
}
