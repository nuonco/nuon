package orgs

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ListWebhooks(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	webhooks, err := s.api.GetCurrentOrgWebhooks(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(webhooks)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"WEBHOOK URL",
			"HAS SECRET",
			"CREATED AT",
		},
	}

	for _, w := range webhooks {
		data = append(data, []string{
			w.ID,
			w.WebhookURL,
			strconv.FormatBool(w.HasSecret),
			w.CreatedAt,
		})
	}

	view.Render(data)
	return nil
}
