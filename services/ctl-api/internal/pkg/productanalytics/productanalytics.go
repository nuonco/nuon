package productanalytics

import (
	"context"

	posthog "github.com/posthog/posthog-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Client struct {
	ph posthog.Client
	l  *zap.Logger
}

func New(lc fx.Lifecycle, cfg *internal.Config, l *zap.Logger) *Client {
	c := &Client{l: l.Named("productanalytics")}
	if cfg.PostHogKey == "" {
		return c
	}

	ph, err := posthog.NewWithConfig(cfg.PostHogKey, posthog.Config{
		Endpoint: cfg.PostHogHost,
		BeforeSend: func(msg posthog.Message) posthog.Message {
			if capture, ok := msg.(posthog.Capture); ok {
				delete(capture.Properties, "$mcp_error_message")
				return capture
			}
			return msg
		},
	})
	if err != nil {
		c.l.Error("unable to create posthog client, product analytics disabled", zap.Error(err))
		return c
	}
	c.ph = ph

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			if err := ph.Close(); err != nil {
				c.l.Warn("unable to flush posthog events", zap.Error(err))
			}
			return nil
		},
	})
	return c
}

func (c *Client) PostHog() posthog.Client {
	if c == nil {
		return nil
	}
	return c.ph
}

func (c *Client) Capture(distinctID, event, orgID string, props posthog.Properties) {
	if c.PostHog() == nil || distinctID == "" {
		return
	}

	msg := posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: props,
	}
	if orgID != "" {
		msg.Groups = posthog.NewGroups().Set("organization", orgID)
	}
	if err := c.ph.Enqueue(msg); err != nil {
		c.l.Warn("unable to enqueue posthog event", zap.String("event", event), zap.Error(err))
	}
}
