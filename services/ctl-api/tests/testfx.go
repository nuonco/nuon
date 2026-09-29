package tests

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v50/github"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/salesforce"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/autolink"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
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

type TestMocks struct {
	MockTC temporalclient.Client
	MockGH vcshelpers.GithubClient
	MockTF terraform.Client
}

type TestOpts struct {
	T               testing.TB
	Mocks           *TestMocks
	CustomValidator bool
}

func CtlApiFXOptions(t testing.TB) []fx.Option {
	return CtlApiFXOptionsWithMocks(TestOpts{T: t, CustomValidator: true})
}

// Deprecated: Use CtlApiFXOptionsWithMocks(tests.TestOpts{}) instead.
func CtlApiFXOptionsWithValidator(t testing.TB) []fx.Option {
	return CtlApiFXOptionsWithMocks(TestOpts{T: t, CustomValidator: false})
}

func CtlApiFXOptionsWithMocks(opts TestOpts) []fx.Option {
	options := []fx.Option{
		fx.WithLogger(NopFxLogger),

		fx.Provide(internal.NewConfig),
		fx.Provide(telemetry.NewConfig),

		fx.Provide(log.New),
		fx.Provide(dblog.New),

		fx.Provide(loops.New),
		fx.Provide(salesforce.New),
		fx.Provide(func() *github.Client { return github.NewClient(nil) }),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(features.New),

		fx.Provide(blobstore.NewService),

		fx.Provide(func() *slackclient.Client { return slackclient.New() }),
		fx.Provide(autolink.New),

		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(blob.AsBlob(blob.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),

		fx.Provide(func(cfg *internal.Config) *querycollector.Collector {
			if cfg.DebugEnableQueryCollector {
				return querycollector.NewCollector(5000)
			}
			return nil
		}),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		fx.Provide(authz.New),
		fx.Provide(analytics.New),
		fx.Provide(account.New),

		// why: Queue client (uses mock temporal client). The enqueuer gets a no-op
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

		fx.Provide(emitterclient.New),

		fx.Provide(accountshelpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(runnershelpers.New),
		fx.Provide(runbookshelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(orgshelpers.New),

		fx.Provide(api.NewEndpointAudit),
		fx.Provide(audit.New),

		fx.Provide(testseed.New),

		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
	}

	if opts.CustomValidator {
		options = append(options, fx.Provide(validatorpkg.New))
	} else {
		options = append(options, fx.Provide(validator.New))
	}

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
