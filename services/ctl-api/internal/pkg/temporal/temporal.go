package temporal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	L          *zap.Logger
	V          *validator.Validate
	Cfg        *internal.Config
	MW         metrics.Writer
	Propagator workflow.ContextPropagator
}

func New(params Params) (temporalclient.Client, error) {
	dataConverter := dataconverter.NewJSONConverter()

	tc, err := temporalclient.New(params.V,
		temporalclient.WithAddr(params.Cfg.TemporalHost),
		temporalclient.WithLogger(params.L),
		temporalclient.WithNamespace(params.Cfg.TemporalNamespace),
		temporalclient.WithDataConverter(dataConverter),
		temporalclient.WithContextPropagator(params.Propagator),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	return tc, nil
}
