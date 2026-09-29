package service

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func newTelemetryRelayEndpoint(cfg *internal.Config, issuer *telemetryTokenIssuer) (string, error) {
	if cfg == nil || cfg.TelemetryRelayEndpoint == "" {
		return "", nil
	}
	if issuer == nil {
		return "", fmt.Errorf("telemetry relay endpoint requires a configured telemetry token issuer")
	}
	if err := app.ValidateTelemetryRelayEndpoint(cfg.TelemetryRelayEndpoint); err != nil {
		return "", err
	}
	return cfg.TelemetryRelayEndpoint, nil
}
