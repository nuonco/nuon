package temporal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func New(l *zap.Logger, v *validator.Validate, cfg *internal.Config) (temporalclient.Client, error) {
	dataConverter := dataconverter.NewJSONConverter()

	tc, err := temporalclient.New(v,
		temporalclient.WithAddr(cfg.TemporalHost),
		temporalclient.WithLogger(l),
		temporalclient.WithNamespace(cfg.TemporalNamespace),
		temporalclient.WithDataConverter(dataConverter),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	return tc, nil
}
