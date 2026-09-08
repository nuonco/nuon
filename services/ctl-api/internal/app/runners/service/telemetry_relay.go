package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const maxTelemetryRelayEndpointSize = 4096

func newTelemetryRelayEndpoint(cfg *internal.Config, issuer *telemetryTokenIssuer) (string, error) {
	if cfg == nil || cfg.TelemetryRelayEndpoint == "" {
		return "", nil
	}
	if issuer == nil {
		return "", fmt.Errorf("telemetry relay endpoint requires a configured telemetry token issuer")
	}
	if len(cfg.TelemetryRelayEndpoint) > maxTelemetryRelayEndpointSize {
		return "", fmt.Errorf("telemetry relay endpoint exceeds maximum size")
	}

	endpoint, err := url.Parse(cfg.TelemetryRelayEndpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || strings.Contains(cfg.TelemetryRelayEndpoint, "${") {
		return "", fmt.Errorf("telemetry relay endpoint must be an HTTPS URL with a host and no userinfo, query, fragment, or environment expansion")
	}
	return cfg.TelemetryRelayEndpoint, nil
}
