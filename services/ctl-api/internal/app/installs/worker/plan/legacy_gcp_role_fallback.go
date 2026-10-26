package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

const legacyGCPRoleFallbackVersion = "legacy-gcp-role-fallback-flag-v1"

func (p *Planner) legacyGCPRoleFallbackEnabled(ctx workflow.Context, l *zap.Logger) bool {
	if workflow.GetVersion(ctx, legacyGCPRoleFallbackVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return false
	}

	enabled, err := activities.AwaitGCPLegacyRoleFallbackEnabled(ctx, activities.GCPLegacyRoleFallbackEnabledRequest{})
	if err != nil {
		l.Warn("unable to check gcp legacy role fallback config", zap.Error(err))
		return false
	}
	return enabled
}
