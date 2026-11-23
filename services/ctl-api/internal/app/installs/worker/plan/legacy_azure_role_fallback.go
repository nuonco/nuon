package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) legacyAzureRoleFallbackEnabled(ctx workflow.Context, l *zap.Logger) bool {
	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureAzureLegacyRoleFallback))
	if err != nil {
		l.Warn("unable to check azure legacy role fallback feature", zap.Error(err))
		return false
	}
	return enabled
}
