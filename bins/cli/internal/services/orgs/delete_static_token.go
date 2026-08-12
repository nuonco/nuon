package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeleteStaticToken(ctx context.Context, tokenID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	if tokenID == "" {
		return ui.PrintError(fmt.Errorf("token id is required"))
	}

	if asJSON {
		err := s.api.DeleteStaticToken(ctx, tokenID)
		if err != nil {
			return err
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		ui.PrintJSON(response{
			ID:      tokenID,
			Deleted: true,
		})
		return nil
	}

	view := ui.NewDeleteView("api token", tokenID, s.cfg.Interactive)
	view.Start()
	if err := s.api.DeleteStaticToken(ctx, tokenID); err != nil {
		return view.Fail(err)
	}

	view.Success()
	return nil
}
