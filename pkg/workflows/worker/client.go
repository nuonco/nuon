package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
)

func (w *worker) getClient() (client.Client, func(), error) {
	l, err := w.getLogger()
	if err != nil {
		return nil, nil, err
	}

	dataConverter := dataconverter.NewJSONConverter()
	tc, err := temporalclient.New(w.v,
		temporalclient.WithAddr(w.Config.TemporalHost),
		temporalclient.WithLogger(l),
		temporalclient.WithNamespace(w.Config.TemporalNamespace),
		temporalclient.WithContextPropagators(w.propagators),
		temporalclient.WithDataConverter(dataConverter),
		temporalclient.WithMetricsWriter(w.mw),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	nc, err := tc.GetNamespaceClient(w.Config.TemporalNamespace)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	return nc, func() {
		nc.Close()
		tc.TallyCloser.Close()
	}, nil
}
