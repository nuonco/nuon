package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type mcpWorkflowStarted struct {
	WorkflowID  string `json:"workflow_id"`
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name"`
	Type        string `json:"workflow_type"`
	PlanOnly    bool   `json:"plan_only"`
}

func (s *service) startInstallWorkflow(
	ctx context.Context,
	orgID, installRef string,
	workflowType app.WorkflowType,
	planOnly bool,
	role string,
	metadata map[string]string,
) (*mcpWorkflowStarted, error) {
	install, err := s.findInstall(ctx, orgID, installRef)
	if err != nil {
		return nil, fmt.Errorf("unable to get install %q: %w", installRef, err)
	}

	md := map[string]string{}
	for k, v := range metadata {
		md[k] = v
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx, install.ID, workflowType, md, planOnly, role)
	if err != nil {
		return nil, err
	}

	queueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		return nil, err
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		return nil, fmt.Errorf("enqueue signal: %w", err)
	}

	return &mcpWorkflowStarted{
		WorkflowID:  workflow.ID,
		InstallID:   install.ID,
		InstallName: install.Name,
		Type:        string(workflowType),
		PlanOnly:    planOnly,
	}, nil
}

func mcpRequireDeprovisionConfirm(confirm, planOnly bool) error {
	if planOnly || confirm {
		return nil
	}
	return fmt.Errorf("deprovision requires confirm=true; ask the user to confirm tearing down resources, then retry with confirm set to true")
}
