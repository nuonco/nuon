package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type mcpWorkflowStarted struct {
	WorkflowID  string         `json:"workflow_id"`
	InstallID   string         `json:"install_id"`
	InstallName string         `json:"install_name"`
	Type        string         `json:"workflow_type"`
	PlanOnly    bool           `json:"plan_only"`
	NextAction  *mcpNextAction `json:"next_action,omitempty"`
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
	if authz.HasAuthority(ctx) {
		var operation permissions.PermissionID
		switch workflowType {
		case app.WorkflowTypeDeployComponents:
			operation = permissions.InstallDeploy
		case app.WorkflowTypeDeprovision:
			operation = permissions.InstallDelete
		case app.WorkflowTypeDeprovisionSandbox:
			operation = permissions.InstallTeardown
		case app.WorkflowTypeReprovision, app.WorkflowTypeReprovisionStack, app.WorkflowTypeReprovisionSandbox:
			operation = permissions.InstallProvision
		default:
			return nil, fmt.Errorf("unsupported MCP install workflow")
		}
		if err := s.mcpInstallPoint(ctx, install.ID, operation); err != nil {
			return nil, err
		}
	}

	if err := s.helpers.ValidateInstallRole(ctx, install.ID, role); err != nil {
		return nil, err
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
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		return nil, fmt.Errorf("enqueue signal: %w", err)
	}

	status := string(workflow.Status.Status)
	if status == "" {
		status = string(app.StatusPending)
	}
	return &mcpWorkflowStarted{
		WorkflowID:  workflow.ID,
		InstallID:   install.ID,
		InstallName: install.Name,
		Type:        string(workflowType),
		PlanOnly:    planOnly,
		NextAction:  mcpWatchStart(workflow.ID, status),
	}, nil
}

func mcpRequireDeprovisionConfirm(confirm, planOnly bool) error {
	if planOnly || confirm {
		return nil
	}
	return fmt.Errorf("deprovision requires confirm=true; ask the user to confirm tearing down resources, then retry with confirm set to true")
}
