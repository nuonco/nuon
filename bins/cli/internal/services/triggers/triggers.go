package triggers

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func printTrigger(trigger *models.Trigger, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(trigger)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"id", trigger.ID}, {"name", trigger.Name}, {"preset", trigger.Preset}, {"status", trigger.Status}, {"auth type", trigger.AuthType}, {"envelope", trigger.Envelope}, {"secrets", strconv.Itoa(len(trigger.Secrets))}}))
	return nil
}

func (s *Service) CreateTrigger(ctx context.Context, req *models.TriggerCreateRequest, asJSON bool) error {
	result, err := s.api.CreateTrigger(ctx, req)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"id", result.Trigger.ID}, {"name", result.Trigger.Name}, {"ingress url", result.IngressURL}, {"key id", result.KeyID}, {"secret", result.Secret}, {"setup", triggerSetupGuidance(&result.Trigger)}}))
	return nil
}
func (s *Service) ListTriggers(ctx context.Context, asJSON bool) error {
	triggers, err := s.api.ListTriggers(ctx)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(triggers)
		return nil
	}
	rows := [][]string{{"ID", "NAME", "STATUS", "AUTH TYPE", "LAST EVENT"}}
	for _, trigger := range triggers {
		last := "—"
		if trigger.LastEventAt != nil {
			last = trigger.LastEventAt.String()
		}
		rows = append(rows, []string{trigger.ID, trigger.Name, trigger.Status, trigger.AuthType, last})
	}
	ui.NewListView().Render(sanitizeHumanRows(rows))
	return nil
}
func (s *Service) GetTrigger(ctx context.Context, id string, asJSON bool) error {
	trigger, err := s.api.GetTrigger(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	return printTrigger(trigger, asJSON)
}
func (s *Service) SetTrigger(ctx context.Context, id string, enable, asJSON bool) error {
	var trigger *models.Trigger
	var err error
	if enable {
		trigger, err = s.api.EnableTrigger(ctx, id)
	} else {
		trigger, err = s.api.DisableTrigger(ctx, id)
	}
	if err != nil {
		return ui.PrintError(err)
	}
	return printTrigger(trigger, asJSON)
}
func (s *Service) RotateTriggerSecret(ctx context.Context, id string, asJSON bool) error {
	trigger, err := s.api.GetTrigger(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if trigger.Preset == "slack-events" {
		return ui.PrintError(&ui.CLIUserError{Msg: "Slack signing secrets cannot be rotated; recreate the trigger to replace it"})
	}
	result, err := s.api.RotateTriggerSecret(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"trigger id", result.Trigger.ID}, {"key id", result.KeyID}, {"secret", result.Secret}}))
	return nil
}
func (s *Service) RevokeTriggerSecret(ctx context.Context, triggerID, secretID string, asJSON bool) error {
	result, err := s.api.RevokeTriggerSecret(ctx, triggerID, secretID)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"trigger id", triggerID}, {"secret id", secretID}, {"revoked at", result.RevokedAt.String()}}))
	return nil
}
func (s *Service) DeleteTrigger(ctx context.Context, id string, force, asJSON bool) error {
	if err := s.api.DeleteTrigger(ctx, id, force); err != nil {
		return ui.PrintError(err)
	}
	result := map[string]any{"id": id, "deleted": true}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"id", id}, {"deleted", "true"}}))
	return nil
}

func (s *Service) RevealTriggerIngressURL(ctx context.Context, id string, asJSON bool) error {
	result, err := s.api.GetTriggerIngressURL(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"trigger id", id}, {"ingress url", result.IngressURL}}))
	return nil
}

func (s *Service) ReplaceTriggerIngressURL(ctx context.Context, id string, asJSON bool) error {
	result, err := s.api.RotateTriggerIngressURL(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"trigger id", id}, {"ingress url", result.IngressURL}}))
	return nil
}

func (s *Service) RevealTriggerSecret(ctx context.Context, triggerID, secretID string, asJSON bool) error {
	result, err := s.api.RevealTriggerSecret(ctx, triggerID, secretID)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"trigger id", triggerID}, {"key id", result.KeyID}, {"secret", result.Secret}}))
	return nil
}

func triggerSetupGuidance(trigger *models.Trigger) string {
	switch trigger.Preset {
	case "github", "gitlab", "bitbucket", "gitea", "forgejo", "terraform-cloud", "azure-devops":
		return "Configure the provider webhook URL with the ingress URL and generated secret."
	case "google-pubsub":
		return "Create a Pub/Sub push subscription using the ingress URL and configured OIDC identity."
	case "aws-eventbridge":
		return "Configure an EventBridge API destination using the ingress URL and API key."
	case "aws-sns":
		return "Subscribe the ingress URL to the configured SNS topic."
	case "azure-event-grid":
		return "Create an Event Grid webhook subscription using the ingress URL and API key."
	case "slack-events":
		return "Set the Slack Events API Request URL to the ingress URL. The signing secret is write-only."
	case "datadog":
		return "Create a Datadog Webhooks integration using the ingress URL, API key, and custom payload."
	default:
		return "Send events to the ingress URL using the configured authentication and envelope."
	}
}
func (s *Service) GetDispatch(ctx context.Context, id string, asJSON bool) error {
	result, err := s.api.GetTriggerEventDispatch(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"id", result.ID}, {"status", result.Status}, {"event id", result.TriggerEventID}, {"attempts", strconv.Itoa(result.Attempts)}, {"error", result.Error}}))
	return nil
}
func (s *Service) RetryDispatch(ctx context.Context, id string, asJSON bool) error {
	result, err := s.api.RetryTriggerEventDispatch(ctx, id)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render(sanitizeHumanRows([][]string{{"dispatch id", result.DispatchID}, {"retry id", result.RetryID}}))
	return nil
}
