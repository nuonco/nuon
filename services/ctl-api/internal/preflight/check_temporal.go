package preflight

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	tclient "go.temporal.io/sdk/client"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

var temporalCheck = Check{
	Name:        "temporal",
	Description: "temporal server health",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "temporal_host", Value: cfg.TemporalHost, Required: true},
			{Name: "temporal_namespace", Value: cfg.TemporalNamespace},
		}
	},

	// Dials and calls CheckHealth rather than opening a TCP socket: a listening
	// port proves nothing about whether the frontend service is actually serving.
	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		client, err := temporalclient.New(validator.New(),
			temporalclient.WithAddr(cfg.TemporalHost),
			temporalclient.WithNamespace(cfg.TemporalNamespace),
			temporalclient.WithLogger(nopLogger()),
		)
		if err != nil {
			return "", fmt.Errorf("dial failed: %w", err)
		}
		defer client.Close()

		if _, err := client.CheckHealth(ctx, &tclient.CheckHealthRequest{}); err != nil {
			return "", fmt.Errorf("health check failed: %w", err)
		}

		return fmt.Sprintf("healthy %s",
			summary("host", cfg.TemporalHost, "namespace", cfg.TemporalNamespace)), nil
	},
}
