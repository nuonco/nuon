package cmd

import (
	"github.com/nuonco/nuon/sdks/nuon-go"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/cli/internal/services/actions"
	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/auth"
	"github.com/nuonco/nuon/bins/cli/internal/services/builds"
	"github.com/nuonco/nuon/bins/cli/internal/services/components"
	"github.com/nuonco/nuon/bins/cli/internal/services/docs"
	"github.com/nuonco/nuon/bins/cli/internal/services/installs"
	"github.com/nuonco/nuon/bins/cli/internal/services/orgs"
	"github.com/nuonco/nuon/bins/cli/internal/services/roles"
	"github.com/nuonco/nuon/bins/cli/internal/services/runbooks"
	"github.com/nuonco/nuon/bins/cli/internal/services/secrets"
	"github.com/nuonco/nuon/bins/cli/internal/services/serviceaccounts"
	"github.com/nuonco/nuon/bins/cli/internal/services/triggers"
	"github.com/nuonco/nuon/bins/cli/internal/services/variables"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

// populateDeps wires the CLI's dependencies through fx and populates them onto
// c. Unlike a daemon (see bins/runner), the CLI has no long-running lifecycle:
// the graph is built once per invocation from the pre-run hook — after cobra
// has parsed flags and the auth token has been resolved — and is never
// started.
func (c *cli) populateDeps() error {
	app := fx.New(
		fx.NopLogger,
		fx.Supply(c.v, c.cfg),
		fx.Provide(
			newAPIClient,
			actions.New,
			apps.New,
			auth.New,
			builds.New,
			components.New,
			docs.New,
			installs.New,
			orgs.New,
			roles.New,
			runbooks.New,
			secrets.New,
			serviceaccounts.New,
			variables.New,
			version.New,
			func(api nuon.Client) *triggers.Service { return triggers.New(api) },
		),
		fx.Populate(
			&c.apiClient,
			&c.actions,
			&c.apps,
			&c.auth,
			&c.builds,
			&c.components,
			&c.docs,
			&c.installs,
			&c.orgs,
			&c.roles,
			&c.runbooks,
			&c.secrets,
			&c.serviceAccounts,
			&c.triggers,
			&c.variables,
			&c.version,
		),
	)

	return app.Err()
}
