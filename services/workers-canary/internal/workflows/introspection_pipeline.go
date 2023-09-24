package workflows

import (
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) execIntrospection(ctx workflow.Context, outputs *activities.TerraformRunOutputs) error {
	l := workflow.GetLogger(ctx)
	if w.cfg.DisableIntrospection {
		l.Info("skipping cli commands due to local config", zap.String("env", "DISABLE_INTROSPECTION"))
		return nil
	}

	return nil
}
