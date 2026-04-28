package orgs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateWebhook(ctx context.Context, webhookURL, webhookSecret string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if webhookURL == "" {
		return view.Error(fmt.Errorf("webhook url is required"))
	}

	webhook, err := s.api.CreateCurrentOrgWebhook(ctx, &models.ServiceCreateCurrentOrgWebhookRequest{
		WebhookURL:    &webhookURL,
		WebhookSecret: webhookSecret,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(webhook)
		return nil
	}

	view.Render([][]string{
		{"id", webhook.ID},
		{"org id", webhook.OrgID},
		{"webhook url", webhook.WebhookURL},
		{"has secret", strconv.FormatBool(webhook.HasSecret)},
		{"created at", webhook.CreatedAt},
		{"created by", webhook.CreatedByID},
	})
	return nil
}
