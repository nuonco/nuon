package workflows

import (
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("deprovisioning")
	resp := &canaryv1.DeprovisionResponse{
		CanaryId: req.CanaryId,
		OrgId:    req.OrgId,
	}
	w.sendNotification(ctx, notificationTypeDeprovisionStart, req.CanaryId, nil)

	err := w.execDeprovision(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to deprovision canary: %w", err)
		w.sendNotification(ctx, notificationTypeDeprovisionError, req.CanaryId, err)
		return nil, err
	}

	w.sendNotification(ctx, notificationTypeDeprovisionSuccess, req.CanaryId, nil)
	return resp, nil
}
