package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
)

func (w *worker) getClient() (client.Client, func(), error) {
	l, err := w.getLogger()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get logger: %w", err)
	}

	dataConverter := dataconverter.NewJSONConverter()
	c, err := client.Dial(client.Options{
		HostPort:      w.Config.TemporalHost,
		Namespace:     w.Namespace,
		Logger:        temporalzap.NewLogger(l),
		DataConverter: dataConverter,
	})
	if err != nil {
		l.Error("failed to instantiate temporal client", zap.Error(err))
		return nil, nil, fmt.Errorf("unable to instantiate temporal client: %w", err)
	}

	return c, c.Close, nil
}
