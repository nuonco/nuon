package notifications

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (n *Notifications) Send(ctx context.Context, cfg *app.NotificationsConfig, typ Type, vars map[string]string) error {
	if n.Cfg.DisableNotifications {
		n.MetricsWriter.Incr("notification", metrics.ToTags(map[string]string{
			"status": "noop",
		}))
		n.L.Debug("skipping notification, disabled", zap.String("type", typ.String()))
		return nil
	}

	// handle email notifications
	if cfg.EnableEmailNotifications {
		if err := n.sendEmailNotification(ctx, typ, vars); err != nil {
			n.MetricsWriter.Incr("notification.email", metrics.ToStatusTag("err"))
			return fmt.Errorf("unable to send email notification: %w", err)
		}

		n.MetricsWriter.Incr("notification.email", metrics.ToStatusTag("ok"))
	}
	if !cfg.EnableEmailNotifications {
		n.MetricsWriter.Incr("notification.email", metrics.ToStatusTag("noop"))
	}

	// handle slack notifications
	if cfg.EnableSlackNotifications {
		for _, slackWebhookURL := range cfg.SlackWebhookURLs {
			if err := n.sendSlackNotification(ctx, typ, slackWebhookURL, vars); err != nil {
				n.MetricsWriter.Incr("notification.slack", metrics.ToStatusTag("err"))
				return fmt.Errorf("unable to send slack notification: %w", err)
			}

			n.MetricsWriter.Incr("notification.slack", metrics.ToStatusTag("ok"))
		}
	}
	if !cfg.EnableSlackNotifications {
		n.MetricsWriter.Incr("notification.slack", metrics.ToStatusTag("noop"))
	}

	return nil
}
