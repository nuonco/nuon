package productanalytics

import (
	"context"
	"net/url"
	"strings"

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
		c.l.Debug("posthog capture disabled")
		return c
	}

	env := string(cfg.Env)
	host := apiHost(cfg.PublicAPIURL)
	c.l.Debug("posthog capture config", zap.String("env", env), zap.String("api_host", host))
	ph, err := posthog.NewWithConfig(cfg.PostHogKey, posthog.Config{
		Endpoint:   cfg.PostHogHost,
		BeforeSend: stampDeployment(c.l, env, host),
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

func stampDeployment(l *zap.Logger, env, host string) func(posthog.Message) posthog.Message {
	return func(msg posthog.Message) posthog.Message {
		capture, ok := msg.(posthog.Capture)
		if !ok {
			return msg
		}
		if capture.Properties == nil {
			capture.Properties = posthog.NewProperties()
		}
		for _, key := range []string{"$mcp_error_message", "$mcp_client_user_agent"} {
			delete(capture.Properties, key)
		}
		if env != "" {
			capture.Properties.Set("env", env)
		}
		if host != "" {
			capture.Properties.Set("api_host", host)
		}
		if l != nil {
			l.Debug("posthog capture",
				zap.String("event", capture.Event),
				zap.String("tool", propString(capture.Properties, "$mcp_tool_name")),
				zap.String("agent", propString(capture.Properties, "agent")),
				zap.String("command", propString(capture.Properties, "command")),
				zap.String("env", propString(capture.Properties, "env")),
				zap.String("api_host", propString(capture.Properties, "api_host")),
			)
		}
		return capture
	}
}

func propString(props posthog.Properties, key string) string {
	if props == nil {
		return ""
	}
	value, ok := props[key].(string)
	if !ok {
		return ""
	}
	return value
}

func apiHost(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}
