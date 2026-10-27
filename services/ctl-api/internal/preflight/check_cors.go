package preflight

import (
	"context"
	"strconv"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

var corsCheck = Check{
	Name:        "cors",
	Description: "cors origin restriction",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "cors_allow_all_origins", Value: strconv.FormatBool(cfg.CORSAllowAllOrigins)},
			{Name: "cors_allow_loopback_origins", Value: strconv.FormatBool(cfg.CORSAllowLoopbackOrigins)},
			{Name: "service_deployment", Value: cfg.ServiceDeployment},
			{Name: "app_url", Value: cfg.AppURL},
		}
	},

	Probe: func(_ context.Context, cfg *internal.Config) (string, error) {
		if cfg.CORSAllowAllOrigins {
			return "", warnf("cors_allow_all_origins is set; any origin is reflected with credentials")
		}
		if cfg.CORSAllowLoopbackOrigins || cfg.ServiceDeployment == "local" {
			return "", warnf("plain-http loopback origins are allowed with credentials")
		}

		return "origin restricted to app_url", nil
	},
}
