package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/autolink"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/statejwt"
)

// SlackLibsModule provides the shared Slack helper libraries (Web API client
// and OAuth state JWT encoder) used by the slack service package as well as
// the Slack-listener handlers in Phase 4.
//
// The state-JWT encoder requires the SlackStateJWTSecret config value; in
// dev environments where the secret is unset, statejwt.New() returns an
// error and the surface that depends on it (install URL endpoint) will fail
// to start. That's intentional — install flows must not silently issue
// unsigned state values.
var SlackLibsModule = fx.Module("slack-libs",
	fx.Provide(func() *slackclient.Client {
		return slackclient.New()
	}),
	fx.Provide(func(cfg *internal.Config) (*statejwt.Encoder, error) {
		return statejwt.New(cfg.SlackStateJWTSecret)
	}),
	fx.Provide(autolink.New),
)
