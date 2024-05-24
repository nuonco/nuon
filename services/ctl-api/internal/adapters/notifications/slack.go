package notifications

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/sender"
)

func (n *Notifications) sendSlackNotification(ctx context.Context, typ Type, webhookURL string, vars map[string]string) error {
	if typ.SlackNotificationTemplate() == "" {
		return nil
	}

	ss, err := sender.NewSlackSender(webhookURL)
	if err != nil {
		return fmt.Errorf("unable to create slack sender: %w", err)
	}

	msg, err := sender.TemplateMessage(typ.SlackNotificationTemplate(), vars)
	if err != nil {
		return fmt.Errorf("unable to template message: %w", err)
	}

	err = ss.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("unable to send message: %w", err)
	}

	return nil
}
