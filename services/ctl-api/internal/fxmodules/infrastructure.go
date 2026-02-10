package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	pkglog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/s3payload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// InfrastructureModule provides all core infrastructure dependencies
// including config, logging, databases, temporal, and other shared services.
var InfrastructureModule = fx.Module("infrastructure",
	// Config and logging foundation
	fx.Provide(internal.NewConfig),
	fx.WithLogger(pkglog.NewFXLog),
	fx.Provide(log.New),
	fx.Provide(dblog.New),

	// Database connections
	fx.Provide(psql.AsPSQL(psql.New)),
	fx.Provide(ch.AsCH(ch.New)),

	// Blob storage service
	fx.Provide(blobstore.NewService),

	// Temporal data converters and client
	fx.Provide(gzip.AsGzip(gzip.New)),
	fx.Provide(largepayload.AsLargePayload(largepayload.New)),
	fx.Provide(s3payload.AsS3Payload(s3payload.New)),
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
	fx.Provide(eventloop.New),
	fx.Provide(teventloop.New),
	fx.Provide(authz.New),
	fx.Provide(features.New),
	fx.Provide(account.New),
	fx.Provide(analytics.New),
	fx.Provide(analytics.NewTemporal),
	fx.Provide(cloudformation.NewTemplates),
)
