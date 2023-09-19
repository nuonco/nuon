package workflows

import (
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("deprovisioning")
	//if err := req.Validate(); err != nil {
		//return nil, err
	//}

	resp := &canaryv1.DeprovisionResponse{
		CanaryId: req.CanaryId,
		OrgId:	  req.OrgId,
	}
	w.sendNotification(ctx, notificationTypeDeprovisionStart, req.CanaryId, nil)

	w.sendNotification(ctx, notificationTypeDeprovisionSuccess, req.CanaryId, nil)
	return resp, nil
}
