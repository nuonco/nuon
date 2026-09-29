package app

import (
	"fmt"
	"net/url"
	"strings"
)

func (s OrgTelemetrySettings) ResolveRelayEndpoint(deploymentEndpoint string) string {
	if s.RelayEndpoint != nil {
		return *s.RelayEndpoint
	}
	return deploymentEndpoint
}

func ValidateTelemetryRelayEndpoint(value string) error {
	if len(value) > 4096 {
		return fmt.Errorf("telemetry relay endpoint exceeds maximum size")
	}
	endpoint, err := url.Parse(value)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.ForceQuery || endpoint.Fragment != "" || strings.ContainsAny(value, "${}\\\t\r\n ") {
		return fmt.Errorf("telemetry relay endpoint must be an HTTPS URL with a host and no userinfo, query, fragment, whitespace, or environment expansion")
	}
	return nil
}
