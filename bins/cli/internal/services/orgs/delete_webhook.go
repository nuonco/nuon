package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeleteWebhook(ctx context.Context, webhookID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	if webhookID == "" {
		return ui.PrintError(fmt.Errorf("webhook id is required"))
	}

	if asJSON {
		err := s.api.DeleteCurrentOrgWebhook(ctx, webhookID)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		ui.PrintJSON(response{
			ID:      webhookID,
			Deleted: true,
		})
		return nil
	}

	view := ui.NewDeleteView("webhook", webhookID, s.cfg.Interactive)
	view.Start()
	if err := s.api.DeleteCurrentOrgWebhook(ctx, webhookID); err != nil {
		return view.Fail(err)
	}

	view.Success()
	return nil
}
