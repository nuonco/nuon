package preflight

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	svcconfig "github.com/nuonco/nuon/pkg/services/config"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

// Dev-only defaults registered in internal/config.go so the slack fx module can
// build without SLACK_* set. Reaching a deployed environment with either still
// in place means signed-webhook verification is trivially forgeable.
const (
	devSlackSigningSecret  = "insecure-slack-signing-secret-for-dev-only"
	devSlackStateJWTSecret = "insecure-slack-state-jwt-secret-for-dev-only"
)

func devPlaceholders(cfg *internal.Config) []string {
	var found []string
	if cfg.SlackSigningSecret == devSlackSigningSecret {
		found = append(found, "slack_signing_secret")
	}
	if cfg.SlackStateJWTSecret == devSlackStateJWTSecret {
		found = append(found, "slack_state_jwt_secret")
	}

	return found
}

var slackCheck = Check{
	Name:        "slack",
	Description: "slack app configuration",

	// The whole integration is optional: no client id means no slack app, which
	// is a supported deployment rather than a misconfiguration.
	Skip: func(cfg *internal.Config) (string, bool) {
		if cfg.SlackClientID == "" {
			return "slack app not configured", true
		}

		return "", false
	},

	// Only the two values with no registered default can actually be missing;
	// the signing and state secrets always hold at least their dev default, so
	// the placeholder comparison in Probe is what covers them.
	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "slack_client_id", Value: cfg.SlackClientID},
			{Name: "slack_client_secret", Value: cfg.SlackClientSecret, Required: true, Secret: true},
			{Name: "slack_signing_secret", Value: cfg.SlackSigningSecret, Secret: true},
			{Name: "slack_state_jwt_secret", Value: cfg.SlackStateJWTSecret, Secret: true},
			{Name: "slack_oauth_redirect_url", Value: cfg.SlackOAuthRedirectURL, Required: true},
			{Name: "slack_http_port", Value: cfg.SlackHTTPPort},
		}
	},

	// No live Slack call: every Slack API method needs a per-workspace token
	// held in the database, not in config. What config alone can be wrong about
	// is placeholder secrets and a malformed redirect URL.
	Probe: func(_ context.Context, cfg *internal.Config) (string, error) {
		redirect, err := url.Parse(cfg.SlackOAuthRedirectURL)
		if err != nil {
			return "", fmt.Errorf("slack_oauth_redirect_url is not a valid URL: %w", err)
		}
		if redirect.Scheme != "https" || redirect.Host == "" {
			return "", fmt.Errorf("slack_oauth_redirect_url must be an absolute https URL, got %q",
				cfg.SlackOAuthRedirectURL)
		}

		// Production supplies these explicitly, so a value still equal to the
		// registered dev default there means webhook signature verification is
		// forgeable. Everywhere else the default is the intended value, and a
		// laptop reports whatever env its configmap carries — so warn rather
		// than fail outside production.
		if placeholders := devPlaceholders(cfg); len(placeholders) > 0 {
			joined := strings.Join(placeholders, ", ")
			if cfg.Env == svcconfig.Production {
				return "", fmt.Errorf("insecure dev default in production for %s", joined)
			}

			return "", warnf("insecure dev default still in place for %s %s",
				joined, summary("env", cfg.Env.String()))
		}

		// Same shape as the google/github auth providers: everything config can
		// be wrong about is checked, but nothing confirms Slack accepts it.
		return "", warnf("config valid but unverified: no live API call without a workspace token %s",
			summary("client_id", cfg.SlackClientID, "redirect", redirect.Host))
	},
}
