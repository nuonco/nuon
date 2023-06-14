package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

func (w *worker) getClient() (client.Client, func(), error) {
	l, err := w.getLogger()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get logger: %w", err)
	}

	c, err := client.Dial(client.Options{
		HostPort:  w.Config.TemporalHost,
		Namespace: w.Namespace,
		Logger:    temporalzap.NewLogger(l),
	})
	if err != nil {
		l.Error("failed to instantiate temporal client", zap.Error(err))
		return nil, nil, fmt.Errorf("unable to instantiate temporal client: %w", err)
	}

	return c, c.Close, nil
}
