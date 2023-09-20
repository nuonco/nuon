package workflows

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) Provision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.ProvisionResponse, error) {
	resp := &canaryv1.ProvisionResponse{
		CanaryId: req.CanaryId,
	}

	l := workflow.GetLogger(ctx)
	l.Info("provisioning canary", "id", req.CanaryId)

	ensureCanaryID := workflow.SideEffect(ctx, func(_ workflow.Context) interface{} {
		if req.CanaryId != "" {
			return req.CanaryId
		}

		newCanaryID := domains.NewCanaryID()
		return newCanaryID
	})
	var canaryID string
	if err := ensureCanaryID.Get(&canaryID); err != nil {
		return resp, fmt.Errorf("unable to get canary ID: %w", err)
	}
	req.CanaryId = canaryID

	if err := req.Validate(); err != nil {
		return resp, err
	}

	w.sendNotification(ctx, notificationTypeProvisionStart, req.CanaryId, nil)
	outputs, err := w.execProvision(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to provision canary: %w", err)
		w.sendNotification(ctx, notificationTypeProvisionError, req.CanaryId, err)
		return nil,err
	}
	w.sendNotification(ctx, notificationTypeSuccess, req.CanaryId, nil)

	err = w.execCLICommands(ctx, outputs)
	if err != nil {
		err = fmt.Errorf("unable to execute cli commands: %w", err)
		w.sendNotification(ctx, notificationTypeProvisionError, req.CanaryId, err)
	}
	w.sendNotification(ctx, notificationTypeSuccess, req.CanaryId, nil)

	if err := w.execProvisionDeprovision(ctx,outputs.OrgID, req); err != nil {
		err = fmt.Errorf("unable to start deprovision workflow")
		return resp, err
	}

	return resp, nil
}

func (w *wkflow) execProvisionDeprovision(ctx workflow.Context, orgID string, req *canaryv1.ProvisionRequest) error {
	if !req.Deprovision {
		return nil
	}
	l := workflow.GetLogger(ctx)

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Hour * 24,
		WorkflowTaskTimeout:	  time.Hour,
		TaskQueue:		  workflows.DefaultTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "Deprovision", &canaryv1.DeprovisionRequest{
		CanaryId: req.CanaryId,
	})

	var resp canaryv1.DeprovisionResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}
	l.Debug("deprovision response", "response", &resp)
	return nil
}
