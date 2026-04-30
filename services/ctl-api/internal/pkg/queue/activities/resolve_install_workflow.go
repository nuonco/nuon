package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// ResolveInstallWorkflowResult is the result of resolving the install workflow
// that owns a given install_workflow_step. WorkflowID and WorkflowType are
// empty strings when no workflow could be resolved (best-effort lookup).
type ResolveInstallWorkflowResult struct {
	WorkflowID   string `json:"workflow_id"`
	WorkflowType string `json:"workflow_type"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field workflowStepID
func (a *Activities) resolveInstallWorkflowByStepID(ctx context.Context, workflowStepID string) (*ResolveInstallWorkflowResult, error) {
	if workflowStepID == "" {
		return &ResolveInstallWorkflowResult{}, nil
	}

	stepTable := (&app.WorkflowStep{}).TableName()
	wfTable := (&app.Workflow{}).TableName()

	var row ResolveInstallWorkflowResult
	res := a.db.WithContext(ctx).
		Table(stepTable).
		Select(wfTable+".id AS workflow_id, "+wfTable+".type AS workflow_type").
		Joins("JOIN "+wfTable+" ON "+wfTable+".id = "+stepTable+".install_workflow_id").
		Where(stepTable+".id = ?", workflowStepID).
		Take(&row)
	if res.Error != nil {
		return &ResolveInstallWorkflowResult{}, generics.TemporalGormError(res.Error, "unable to resolve install workflow")
	}

	return &row, nil
}
