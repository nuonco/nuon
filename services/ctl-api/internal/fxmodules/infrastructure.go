package fxmodules

import (
	"context"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	notebookclient "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	pkglog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/arm"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/blob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type queryWriterParams struct {
	fx.In
	Cfg       *internal.Config
	Collector *querycollector.Collector
	CHDB      *gorm.DB `name:"ch"`
	L         *zap.Logger
}

// InfrastructureModule provides all core infrastructure dependencies
// including config, logging, databases, temporal, and other shared services.
var InfrastructureModule = fx.Module("infrastructure",
	// Config and logging foundation
	fx.Provide(internal.NewConfig),
	fx.WithLogger(pkglog.NewFXLog),
	fx.Provide(log.New),
	fx.Provide(dblog.New),

	// Query collector (enabled by debug_enable_query_collector config)
	fx.Provide(func(cfg *internal.Config) *querycollector.Collector {
		if cfg.DebugEnableQueryCollector {
			return querycollector.NewCollector(5000)
		}
		return nil
	}),

	// Database connections
	fx.Provide(psql.AsPSQL(psql.New)),
	fx.Provide(ch.AsCH(ch.New)),

	// Query collector ClickHouse writer (optional, writes captured queries to CH)
	fx.Invoke(func(lc fx.Lifecycle, p queryWriterParams) {
		if p.Collector == nil {
			return
		}

		disabledTables := make(map[string]struct{})
		if p.Cfg.QueryCollectorDisabledTables != "" {
			for _, t := range strings.Split(p.Cfg.QueryCollectorDisabledTables, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					disabledTables[t] = struct{}{}
				}
			}
		}

		w := querycollector.NewWriter(querycollector.WriterConfig{
			DB:             p.CHDB,
			Logger:         p.L,
			DisabledTables: disabledTables,
		})
		p.Collector.SetWriter(w)

		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				w.Start()
				return nil
			},
			OnStop: func(context.Context) error {
				w.Stop()
				return nil
			},
		})
	}),

	// Blob storage service
	fx.Provide(blobstore.NewService),

	// File cache for blob codec
	fx.Provide(func(cfg *internal.Config, l *zap.Logger) *filecache.FileCache {
		cache, err := filecache.New(filecache.Options{
			Dir:      cfg.TemporalBlobCacheDir,
			MaxCount: cfg.TemporalBlobCacheMaxCount,
			MaxBytes: int64(cfg.TemporalBlobCacheMaxSizeMB) * 1024 * 1024,
		})
		if err != nil {
			l.Warn("failed to create blob cache, caching disabled", zap.Error(err))
			return nil
		}
		return cache
	}),

	// Temporal data converters and client
	fx.Provide(gzip.AsGzip(gzip.New)),
	fx.Provide(largepayload.AsLargePayload(largepayload.New)),
	fx.Provide(blob.AsBlob(blob.New)),
	fx.Provide(signaldb.NewPayloadConverter),
	fx.Provide(dataconverter.New),
	fx.Provide(temporal.New),

	// Core services
	fx.Provide(loops.New),
	fx.Provide(github.New),
	fx.Provide(metrics.New),
	fx.Provide(propagator.New),
	fx.Provide(validator.New),
	fx.Provide(notifications.New),
	fx.Provide(terraform.New),
	fx.Provide(authz.New),
	fx.Provide(features.New),
	fx.Provide(account.New),
	fx.Provide(analytics.New),
	fx.Provide(analytics.NewTemporal),
	fx.Provide(cloudformation.NewTemplates),
	fx.Provide(arm.NewTemplates),
	fx.Provide(enqueuer.New),
	fx.Provide(queueclient.New),
	fx.Provide(emitterclient.New),
	fx.Provide(flowclient.New),
	fx.Provide(notebookclient.New),
)
