package tests

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgmetrics "github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	ghpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/blob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
	validatorpkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

func NopFxLogger() fxevent.Logger { return fxevent.NopLogger }

type noopLifecycle struct{}

func (noopLifecycle) Append(fx.Hook) {}

type testEnqueuerParams struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Cfg     *internal.Config
	TClient temporalclient.Client
	L       *zap.Logger
	MW      pkgmetrics.Writer
}

// TestMocks holds optional mock/fake clients that tests can supply to CtlApiFXOptionsWithMocks.
// Tests create their own mock instances and pass them in; FX registers them as the interface types.
type TestMocks struct {
	MockTC temporalclient.Client
	MockGH vcshelpers.GithubClient
	MockTF terraform.Client
}

// TestOpts configures the FX options for integration tests.
type TestOpts struct {
	// T is required for creating default gomock-based mock clients.
	T testing.TB
	// Mocks to inject. Nil fields use default mocks.
	Mocks *TestMocks
	// CustomValidator uses the custom entity_name validator when true,
	// standard validator when false.
	CustomValidator bool
}

// CtlApiFXOptions returns the common FX options used across all ctl-api integration tests.
// For tests that need mocks, use CtlApiFXOptionsWithMocks instead.
func CtlApiFXOptions(t testing.TB) []fx.Option {
	return CtlApiFXOptionsWithMocks(TestOpts{T: t, CustomValidator: true})
}

// CtlApiFXOptionsWithValidator returns common test options with the standard validator.
//
// Deprecated: Use CtlApiFXOptionsWithMocks(tests.TestOpts{}) instead.
func CtlApiFXOptionsWithValidator(t testing.TB) []fx.Option {
	return CtlApiFXOptionsWithMocks(TestOpts{T: t, CustomValidator: false})
}

// CtlApiFXOptionsWithMocks returns FX options for integration tests with configurable
// mock clients and validator choice.
//
// Usage:
//
//	opts := tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
//	    CustomValidator: true,
//	})
//	app := fxtest.New(t, append(opts, fx.Provide(MyService), fx.Populate(&svc))...)
func CtlApiFXOptionsWithMocks(opts TestOpts) []fx.Option {
	options := []fx.Option{
		// Suppress verbose Fx PROVIDE/INVOKE logs in tests
		fx.WithLogger(NopFxLogger),

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

		// Temporal dependencies
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(blob.AsBlob(blob.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),

		// Databases
		fx.Provide(func(cfg *internal.Config) *querycollector.Collector {
			if cfg.DebugEnableQueryCollector {
				return querycollector.NewCollector(5000)
			}
			return nil
		}),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// Clients and dependencies for account client
		fx.Provide(authz.New),
		fx.Provide(analytics.New),
		fx.Provide(account.New),

		// Queue client (uses mock temporal client). The enqueuer gets a no-op
		// lifecycle so its background workers and sweep workflow never start.
		// NOTE: flowclient is intentionally NOT provided here — it imports
		// executeflow, whose import tree reaches back into packages (e.g.
		// config/syncer) whose tests import this package, creating an import
		// cycle. Suites that need it (installs service) provide it locally.
		fx.Provide(func(p testEnqueuerParams) *enqueuer.Enqueuer {
			return enqueuer.New(enqueuer.Params{
				DB:      p.DB,
				Cfg:     p.Cfg,
				TClient: p.TClient,
				L:       p.L,
				MW:      p.MW,
				LC:      noopLifecycle{},
			})
		}),
		fx.Provide(queueclient.New),

		// Queue emitter client (uses mock temporal client)
		fx.Provide(emitterclient.New),

		// Helpers (order matters due to dependencies)
		fx.Provide(accountshelpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(runnershelpers.New),
		fx.Provide(runbookshelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(orgshelpers.New),

		// Endpoint audit
		fx.Provide(api.NewEndpointAudit),

		// Test fixtures
		fx.Provide(testseed.New),

		// Invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	}

	// Validator choice
	if opts.CustomValidator {
		options = append(options, fx.Provide(validatorpkg.New))
	} else {
		options = append(options, fx.Provide(validator.New))
	}

	// Mock/fake client overrides
	if opts.Mocks != nil && opts.Mocks.MockTC != nil {
		options = append(options, fx.Supply(fx.Annotate(opts.Mocks.MockTC, fx.As(new(temporalclient.Client)))))
	} else if opts.T != nil {
		ctrl := gomock.NewController(opts.T)
		mockTC := temporalclient.NewMockClient(ctrl)
		options = append(options, fx.Supply(fx.Annotate(mockTC, fx.As(new(temporalclient.Client)))))
	}

	if opts.Mocks != nil && opts.Mocks.MockGH != nil {
		options = append(options, fx.Supply(fx.Annotate(opts.Mocks.MockGH, fx.As(new(vcshelpers.GithubClient)))))
	}

	if opts.Mocks != nil && opts.Mocks.MockTF != nil {
		options = append(options, fx.Supply(fx.Annotate(opts.Mocks.MockTF, fx.As(new(terraform.Client)))))
	} else {
		options = append(options, fx.Supply(fx.Annotate(terraform.NewFakeClient(), fx.As(new(terraform.Client)))))
	}

	return options
}
