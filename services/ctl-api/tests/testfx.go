package tests

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	ghpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/s3payload"
	validatorpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// CtlApiFXOptions returns the common FX options used across all ctl-api integration tests.
// This includes configuration, logging, databases, external services, helpers, and clients.
//
// Usage:
//
//	app := fxtest.New(
//	    t,
//	    testfx.CtlApiFXOptions()...,
//	    fx.Provide(MyService), // add your service under test
//	    fx.Populate(&myTestService),
//	)
func CtlApiFXOptions() []fx.Option {
	return []fx.Option{
		// Configuration
		fx.Provide(internal.NewConfig),

		// Logging
		fx.Provide(log.New),
		fx.Provide(dblog.New),

		// External services
		fx.Provide(loops.New),
		fx.Provide(ghpkg.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(features.New),

		// Blob storage service
		fx.Provide(blobstore.NewService),

		// Temporal dependencies (required for eventloop)
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(s3payload.AsS3Payload(s3payload.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),

		// Eventloop client (uses real temporal connection)
		fx.Provide(eventloop.New),

		// Databases
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// Validator (use custom validator with entity_name registration)
		fx.Provide(validatorpkg.New),

		// Clients and dependencies for account client
		fx.Provide(authz.New),
		fx.Provide(analytics.New),
		fx.Provide(account.New),

		// Helpers (order matters due to dependencies)
		fx.Provide(accountshelpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(runnershelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(orgshelpers.New),

		// Endpoint audit
		fx.Provide(api.NewEndpointAudit),

		// Test fixtures
		fx.Provide(testseed.New),

		// Invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	}
}

// CtlApiFXOptionsWithValidator returns common test options but uses the standard validator
// instead of the custom validator. Use this when you don't need custom entity_name validation.
//
// Deprecated: Most tests should use CtlApiFXOptions() with the custom validator.
func CtlApiFXOptionsWithValidator() []fx.Option {
	return []fx.Option{
		// Configuration
		fx.Provide(internal.NewConfig),

		// Logging
		fx.Provide(log.New),
		fx.Provide(dblog.New),

		// External services
		fx.Provide(loops.New),
		fx.Provide(ghpkg.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(features.New),

		// Blob storage service
		fx.Provide(blobstore.NewService),

		// Temporal dependencies (required for eventloop)
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(s3payload.AsS3Payload(s3payload.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),

		// Eventloop client (uses real temporal connection)
		fx.Provide(eventloop.New),

		// Databases
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// Validator (standard validator)
		fx.Provide(validator.New),

		// Clients and dependencies for account client
		fx.Provide(authz.New),
		fx.Provide(analytics.New),
		fx.Provide(account.New),

		// Helpers (order matters due to dependencies)
		fx.Provide(accountshelpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(runnershelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(orgshelpers.New),

		// Endpoint audit
		fx.Provide(api.NewEndpointAudit),

		// Invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	}
}
