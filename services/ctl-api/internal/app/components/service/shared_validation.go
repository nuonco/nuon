package service

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// HealthProbeRequest is one synthetic health check declared on a component.
// Command is an argv array, never a shell string.
type HealthProbeRequest struct {
	Type     string   `json:"type,omitempty"`
	Name     string   `json:"name,omitempty"`
	URL      string   `json:"url,omitempty"`
	Command  []string `json:"command,omitempty"`
	Interval string   `json:"interval,omitempty"`
}

// toConfigHealthProbes maps wire probes onto the parsed-config representation the
// shared builders validate and persist.
func toConfigHealthProbes(probes []HealthProbeRequest) []config.ComponentHealthProbeConfig {
	if len(probes) == 0 {
		return nil
	}
	out := make([]config.ComponentHealthProbeConfig, 0, len(probes))
	for _, probe := range probes {
		out = append(out, config.ComponentHealthProbeConfig{
			Type:    probe.Type,
			Name:    probe.Name,
			URL:     probe.URL,
			Command: probe.Command,
		})
	}
	return out
}

func toConfigOperationRoles(roles map[app.OperationType]*string) []config.EntityOperationRole {
	if len(roles) == 0 {
		return nil
	}
	out := make([]config.EntityOperationRole, 0, len(roles))
	for operation, role := range roles {
		if role == nil {
			continue
		}
		out = append(out, config.EntityOperationRole{
			Operation: config.OperationType(operation),
			RoleName:  *role,
		})
	}
	return out
}
