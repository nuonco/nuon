package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) legacyGCPRoleFallbackEnabled(ctx workflow.Context, l *zap.Logger) bool {
	enabled, err := activities.AwaitGCPLegacyRoleFallbackEnabled(ctx, activities.GCPLegacyRoleFallbackEnabledRequest{})
	if err != nil {
		l.Warn("unable to check gcp legacy role fallback config", zap.Error(err))
		return false
	}
	return enabled
}
