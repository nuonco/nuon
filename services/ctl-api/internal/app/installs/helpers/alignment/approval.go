package alignment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) validateResolvePlan(ctx context.Context, db *gorm.DB, st *installAlignmentState, res *InstallAlignmentResource, approvalID string, allowUnresponded bool) (*alignmentApproval, error) {
	var approval app.WorkflowStepApproval
	err := db.WithContext(ctx).Preload("Response").Preload("InstallWorkflowStep").
		Where(app.WorkflowStepApproval{ID: approvalID, OrgID: st.install.OrgID}).First(&approval).Error
	if err != nil {
		if isNotFound(err) {
			return nil, stderr.ErrNotFound{Err: err, Description: "approved plan not found on this install"}
		}
		return nil, fmt.Errorf("unable to get approved plan: %w", err)
	}
	var wf app.Workflow
	err = db.WithContext(ctx).Where(app.Workflow{ID: approval.InstallWorkflowStep.InstallWorkflowID, OrgID: st.install.OrgID,
		OwnerID: st.install.ID, OwnerType: "installs"}).First(&wf).Error
	if err != nil {
		if isNotFound(err) {
			return nil, stderr.ErrNotFound{Err: err, Description: "approved plan not found on this install"}
		}
		return nil, fmt.Errorf("unable to get approved plan workflow: %w", err)
	}
	invalid := func(description string) error {
		return stderr.ErrUser{Err: fmt.Errorf("approved plan %s: %s", approvalID, description), Description: description}
	}
	switch approval.Type {
	case app.TerraformPlanApprovalType, app.KubernetesManifestApprovalType, app.HelmApprovalApprovalType, app.PulumiApprovalType:
	default:
		return nil, invalid("this approval type cannot be reused to resolve drift")
	}
	if !wf.PlanOnly && wf.FinishedAt.IsZero() {
		return nil, invalid("the plan belongs to a deployment that is still in progress")
	}
	if res.appliedAt != nil && approval.CreatedAt.Before(*res.appliedAt) {
		return nil, invalid("the plan predates the resource's last applied change")
	}
	switch res.Kind {
	case InstallAlignmentResourceKindComponent, InstallAlignmentResourceKindImage:
		if approval.OwnerType != "install_deploys" {
			return nil, invalid("the plan does not belong to this component")
		}
		var deploy app.InstallDeploy
		if err := db.WithContext(ctx).Where(app.InstallDeploy{ID: approval.OwnerID, OrgID: st.install.OrgID,
			InstallComponentID: res.ID}).First(&deploy).Error; err != nil {
			if isNotFound(err) {
				return nil, invalid("the plan does not belong to this component")
			}
			return nil, fmt.Errorf("unable to get plan deploy: %w", err)
		}
		if deploy.ComponentBuildID != res.ExpectedBuildID {
			return nil, invalid("the plan was made for a different build than the selected target")
		}
	case InstallAlignmentResourceKindSandbox:
		if approval.OwnerType != "install_sandbox_runs" {
			return nil, invalid("the plan does not belong to this sandbox")
		}
		var run app.InstallSandboxRun
		if err := db.WithContext(ctx).Where(app.InstallSandboxRun{ID: approval.OwnerID, OrgID: st.install.OrgID,
			InstallID: st.install.ID}).First(&run).Error; err != nil {
			if isNotFound(err) {
				return nil, invalid("the plan does not belong to this sandbox")
			}
			return nil, fmt.Errorf("unable to get plan sandbox run: %w", err)
		}
		if run.InstallSandboxID == nil || *run.InstallSandboxID != res.ID {
			return nil, invalid("the plan does not belong to this sandbox")
		}
		matches := st.expectedSandbox != nil && run.AppSandboxConfigID == st.expectedSandbox.ID
		if !matches && res.Category == InstallAlignmentCategoryStateDriftDetected && st.appliedSandbox != nil {
			matches = sandboxScanMatchesApplied(st, &run)
		}
		if !matches {
			return nil, invalid("the plan was made for a different sandbox config than the selected target")
		}
	default:
		return nil, invalid("plans cannot be reused for this resource")
	}
	if approval.Contents == "" && approval.ContentsBlob == nil {
		return nil, invalid("the plan has no contents to reuse")
	}
	if approval.Response == nil {
		if !allowUnresponded {
			return nil, invalid("the plan has not been approved")
		}
	} else if approval.Response.OrgID != st.install.OrgID || approval.Response.Type != app.WorkflowStepApprovalResponseTypeApprove ||
		strings.HasPrefix(approval.Response.Note, "auto-approved") {
		return nil, invalid("the plan was not explicitly approved")
	}
	return &alignmentApproval{approval: &approval, workflow: &wf}, nil
}

func (s *Helpers) findApprovedResolvePlan(ctx context.Context, st *installAlignmentState, res *InstallAlignmentResource) (*alignmentApproval, error) {
	query := s.db.WithContext(ctx).Model(&app.WorkflowStepApproval{}).
		Joins("JOIN install_workflow_step_approval_responses r ON r.install_workflow_step_approval_id = install_workflow_step_approvals.id AND r.deleted_at = 0").
		Where("r.type = ? AND r.note NOT LIKE ?", app.WorkflowStepApprovalResponseTypeApprove, "auto-approved%").
		Where(app.WorkflowStepApproval{OrgID: st.install.OrgID})
	switch res.Kind {
	case InstallAlignmentResourceKindComponent, InstallAlignmentResourceKindImage:
		query = query.Joins("JOIN install_deploys d ON d.id = install_workflow_step_approvals.owner_id AND d.deleted_at = 0").
			Where(app.WorkflowStepApproval{OwnerType: "install_deploys"}).
			Where("d.install_component_id = ? AND d.component_build_id = ?", res.ID, res.ExpectedBuildID)
	case InstallAlignmentResourceKindSandbox:
		query = query.Joins("JOIN install_sandbox_runs sr ON sr.id = install_workflow_step_approvals.owner_id AND sr.deleted_at = 0").
			Where(app.WorkflowStepApproval{OwnerType: "install_sandbox_runs"}).Where("sr.install_sandbox_id = ?", res.ID)
	default:
		return nil, nil
	}
	if res.appliedAt != nil {
		query = query.Where("install_workflow_step_approvals.created_at >= ?", *res.appliedAt)
	}
	var ids []string
	if err := query.Order("install_workflow_step_approvals.created_at DESC, install_workflow_step_approvals.id DESC").
		Limit(5).Pluck("install_workflow_step_approvals.id", &ids).Error; err != nil {
		return nil, fmt.Errorf("unable to find approved plans: %w", err)
	}
	for _, id := range ids {
		aa, err := s.validateResolvePlan(ctx, s.db, st, res, id, false)
		if err == nil {
			return aa, nil
		}
		var userErr stderr.ErrUser
		var notFound stderr.ErrNotFound
		if !errors.As(err, &userErr) && !errors.As(err, &notFound) {
			return nil, err
		}
	}
	return nil, nil
}

func (s *Helpers) ensureResolvePlanApproved(ctx context.Context, tx *gorm.DB, approvalID string) (*app.WorkflowStepApprovalResponse, error) {
	var locked app.WorkflowStepApproval
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").
		Where(app.WorkflowStepApproval{ID: approvalID}).First(&locked).Error; err != nil {
		return nil, fmt.Errorf("unable to lock approval: %w", err)
	}
	var existing app.WorkflowStepApprovalResponse
	err := tx.WithContext(ctx).Where(app.WorkflowStepApprovalResponse{InstallWorkflowStepApprovalID: approvalID}).First(&existing).Error
	if err == nil {
		if existing.Type != app.WorkflowStepApprovalResponseTypeApprove || strings.HasPrefix(existing.Note, "auto-approved") {
			return nil, stderr.ErrUser{Err: fmt.Errorf("plan %s was not explicitly approved", approvalID), Description: "the plan was not explicitly approved"}
		}
		return nil, nil
	}
	if !isNotFound(err) {
		return nil, fmt.Errorf("unable to get approval response: %w", err)
	}
	response := app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approvalID,
		Type:                          app.WorkflowStepApprovalResponseTypeApprove,
		Note:                          "approved via resolve drift",
	}
	if err := tx.WithContext(ctx).Create(&response).Error; err != nil {
		return nil, fmt.Errorf("unable to create approval response: %w", err)
	}
	return &response, nil
}
