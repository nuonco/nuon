package orgs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// UpdateWebhook calls PATCH /v1/orgs/current/webhooks/{webhook_id} to replace
// the webhook subscription (interests + match) and optionally rotate the
// signing secret.
//
// The interests filter and match predicate are replaced wholesale every
// call: omitting --subscription-json / --subscription-file resets the
// webhook to the implicit default (every event in the org, no scoping).
//
// When webhookSecret is empty, the existing secret is left unchanged. When
// non-empty, the API replaces the stored secret with the provided value.
// Removing a secret entirely (delivering payloads unsigned) is intentionally
// not exposed by the CLI — the SDK's swagger model serializes the field with
// omitempty so an empty string cannot be transmitted; use the dashboard for
// that case.
func (s *Service) UpdateWebhook(
	ctx context.Context,
	webhookID, webhookSecret string,
	subscription SubscriptionFlags,
	asJSON bool,
) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if webhookID == "" {
		return view.Error(fmt.Errorf("webhook id is required"))
	}

	payload, err := resolveSubscription(ctx, s.api, s.cfg.Interactive, subscription)
	if err != nil {
		return view.Error(err)
	}

	req := &models.ServiceUpdateCurrentOrgWebhookRequest{
		Interests:     payload.Interests,
		Match:         payload.Match,
		WebhookSecret: webhookSecret,
	}

	webhook, err := s.api.UpdateCurrentOrgWebhook(ctx, webhookID, req)
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
		{"updated at", webhook.UpdatedAt},
	})
	return nil
}
