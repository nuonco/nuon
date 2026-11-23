package fxmodules

import (
	"go.uber.org/zap"

	pkgmetrics "github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func reportURLPolicyWarnings(cfg *internal.Config, l *zap.Logger, mw pkgmetrics.Writer) {
	cfg.URLPolicy.OnWarn(func(violations urlpolicy.Violations) {
		for _, violation := range violations {
			l.Warn("app config value is outside the allowed hosts",
				zap.String("category", string(violation.Category)),
				zap.String("field", violation.Field),
				zap.String("value", violation.Value),
				zap.String("reason", violation.Reason),
			)
			mw.Incr("allowed_hosts.violation", pkgmetrics.ToTags(map[string]string{
				"category": string(violation.Category),
				"mode":     string(urlpolicy.ModeWarn),
			}))
		}
	})
}
