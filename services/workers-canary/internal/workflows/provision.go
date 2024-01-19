package workflows

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *wkflow) getCanaryID(ctx workflow.Context, req *canaryv1.ProvisionRequest) (string, error) {
	ensureCanaryID := workflow.SideEffect(ctx, func(_ workflow.Context) interface{} {
		if req.CanaryId != "" {
			return req.CanaryId
		}

		newCanaryID := domains.NewCanaryID()
		return newCanaryID
	})
	var canaryID string
	if err := ensureCanaryID.Get(&canaryID); err != nil {
		return "", fmt.Errorf("unable to get canary ID: %w", err)
	}

	return canaryID, nil
}

func (w *wkflow) Provision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.ProvisionResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("provisioning canary", "id", req.CanaryId)

	canaryID, err := w.getCanaryID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to get canary ID: %w", err)
	}
	req.CanaryId = canaryID

	if err := req.Validate(); err != nil {
		return nil, err
	}

	w.sendNotification(ctx, notificationTypeCanaryStart, req.CanaryId, req.SandboxMode, nil)
	outputs, orgID, apiToken, err := w.execProvision(ctx, req)
	if err != nil {
		w.sendNotification(ctx, notificationTypeProvisionError, req.CanaryId, req.SandboxMode, err)

		if err := w.execProvisionDeprovision(ctx, orgID, req); err != nil {
			l.Error("unable to deprovision", zap.Error(err))
		}
		return nil, err
	}

	err = w.execTests(ctx, req, outputs, orgID, apiToken)
	if err != nil {
		w.sendNotification(ctx, notificationTypeTestsError, req.CanaryId, req.SandboxMode, err)

		if err := w.execProvisionDeprovision(ctx, orgID, req); err != nil {
			l.Error("unable to deprovision", zap.Error(err))
		}
		return nil, err
	}

	if err := w.execProvisionDeprovision(ctx, orgID, req); err != nil {
		l.Error("unable to deprovision", zap.Error(err))
	}

	w.sendNotification(ctx, notificationTypeCanarySuccess, req.CanaryId, req.SandboxMode, nil)
	return &canaryv1.ProvisionResponse{
		CanaryId: req.CanaryId,
		OrgId:    orgID,
	}, nil
}

func (w *wkflow) execProvisionDeprovision(ctx workflow.Context, orgID string, req *canaryv1.ProvisionRequest) error {
	l := workflow.GetLogger(ctx)
	if orgID == "" {
		l.Info("unable to cleanup, no org id present")
		return nil
	}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-deprovision", req.CanaryId),
		WorkflowExecutionTimeout: time.Hour * 24,
		WorkflowTaskTimeout:      time.Hour,
		TaskQueue:                workflows.DefaultTaskQueue,
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
