package fxmodules

import (
	"go.uber.org/fx"

	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
)

// DatadogLibsModule provides the shared Datadog HTTP client used by the
// signal lifecycle hook and the datadog service handlers.
//
// Unlike Slack, the DD client carries no per-tenant state — credentials
// (API key, application key) live on app.DatadogConnection rows and are
// passed into every call. This means we can register a single shared
// client even when no DD connections exist; the hook's Supports() check
// guards against doing work for orgs without connections.
//
// Wired into the same surfaces as SlackLibsModule (services.go for HTTP
// handlers, worker.go for the lifecycle hook running inside activities).
var DatadogLibsModule = fx.Module("datadog-libs",
	fx.Provide(func() *ddclient.Client {
		return ddclient.New()
	}),
)
