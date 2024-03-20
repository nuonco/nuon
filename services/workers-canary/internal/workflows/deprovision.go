package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/metrics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("deprovisioning")
	resp := &canaryv1.DeprovisionResponse{
		CanaryId: req.CanaryId,
	}

	err := w.execDeprovision(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to deprovision canary: %w", err)
		w.sendNotification(ctx, notificationTypeDeprovisionError, req.CanaryId, req.SandboxMode, err)
		return nil, err
	}

	w.metricsWriter.Incr(ctx, "deprovision", 1, "status:ok", metrics.ToBoolTag("sandbox_mode", req.SandboxMode))
	return resp, nil
}
