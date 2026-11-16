package alignment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
)

var alignmentComponentSignalTypes = []string{
	string(componentdeployapplyplan.SignalType),
	string(componentdeploysyncandplan.SignalType),
	string(componentsyncimage.SignalType),
	string(componentteardownapplyplan.SignalType),
	string(componentteardownsyncandplan.SignalType),
}

func (s *Helpers) explainInstallAlignment(ctx context.Context, st *installAlignmentState, alignment *InstallAlignment) error {
	var resourceIDs, workflowIDs []string
	var since *time.Time
	unboundedHistory := false
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		switch res.Category {
		case InstallAlignmentCategoryChangesNotApplied:
			resourceIDs = append(resourceIDs, res.ID)
			if res.appliedAt == nil {
				unboundedHistory = true
			} else if since == nil || res.appliedAt.Before(*since) {
				since = res.appliedAt
			}
			if res.LastAttempt != nil && res.LastAttempt.WorkflowID != "" {
				workflowIDs = append(workflowIDs, res.LastAttempt.WorkflowID)
			}
			if res.OpenDeployment != nil {
				workflowIDs = append(workflowIDs, res.OpenDeployment.ID)
			}
		case InstallAlignmentCategoryStateDriftDetected:
			res.Explanation = &InstallAlignmentExplanation{Summary: "A drift check found differences between live resources and what was last applied."}
			if res.Plan != nil && res.Plan.PlanOnly {
				res.Explanation.WorkflowID = res.Plan.WorkflowID
				res.Explanation.StepID = res.Plan.StepID
			}
		}
	}
	if len(resourceIDs) == 0 {
		return nil
	}
	var sandboxID, stackID string
	if st.sandbox != nil {
		sandboxID = st.sandbox.ID
	}
	if st.stack != nil {
		stackID = st.stack.ID
	}
	var sandboxSignals, stackSignals []string
	for t := range alignmentSandboxSignalTypes {
		sandboxSignals = append(sandboxSignals, t)
	}
	for t := range alignmentStackSignalTypes {
		stackSignals = append(stackSignals, t)
	}
	allSignals := append(append(append([]string{}, alignmentComponentSignalTypes...), sandboxSignals...), stackSignals...)
	args := []any{sandboxSignals, sandboxID, stackSignals, stackID, st.install.AppID,
		st.install.ID, st.install.OrgID, allSignals}
	historyFilter := ""
	if !unboundedHistory && since != nil {
		historyFilter = "AND (w.created_at >= ? OR w.id IN ?)"
		args = append(args, *since, workflowIDs)
	}
	args = append(args, resourceIDs)
	var latest []struct {
		ResourceID string
		WorkflowID string
	}
	if err := s.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resource_id) resource_id, workflow_id, created_at
		FROM (
			SELECT CASE
				WHEN s.queue_signal->>'type' IN ? OR s.step_target_type IN ('install_sandbox_runs', 'install_sandbox_run') THEN NULLIF(?, '')
				WHEN s.queue_signal->>'type' IN ? OR s.step_target_type IN ('install_stack_versions', 'install_stack_version') THEN NULLIF(?, '')
				ELSE ic.id END AS resource_id, w.id AS workflow_id, w.created_at
			FROM install_workflows w
			JOIN install_workflow_steps s ON s.install_workflow_id = w.id AND s.deleted_at = 0 AND NOT s.retried
			LEFT JOIN install_deploys d ON d.id = s.step_target_id AND d.deleted_at = 0
			LEFT JOIN components c ON c.name = s.metadata->'component_name' AND c.app_id = ? AND c.deleted_at = 0
			LEFT JOIN install_components ic ON ic.install_id = w.owner_id AND ic.deleted_at = 0
				AND (ic.component_id = s.queue_signal->'data'->>'ComponentID' OR ic.component_id = c.id OR ic.id = d.install_component_id)
			WHERE w.owner_id = ? AND w.owner_type = 'installs' AND w.org_id = ? AND w.deleted_at = 0 AND NOT w.plan_only
				AND (s.queue_signal->>'type' IN ? OR s.step_target_type IN ('install_deploys', 'install_deploy', 'install_sandbox_runs', 'install_sandbox_run', 'install_stack_versions', 'install_stack_version'))
				`+historyFilter+`
		) matched
		WHERE resource_id IN ?
		ORDER BY resource_id, created_at DESC, workflow_id DESC`,
		args...).Scan(&latest).Error; err != nil {
		return fmt.Errorf("unable to get latest resource workflows: %w", err)
	}
	byResource := map[string]string{}
	for _, row := range latest {
		byResource[row.ResourceID] = row.WorkflowID
	}
	for _, res := range alignment.Resources {
		if res.Category != InstallAlignmentCategoryChangesNotApplied {
			continue
		}
		if res.OpenDeployment != nil {
			byResource[res.ID] = res.OpenDeployment.ID
		}
		if id := byResource[res.ID]; id != "" {
			workflowIDs = append(workflowIDs, id)
		}
	}
	workflowsByID := map[string]*app.Workflow{}
	stepsByWorkflow := map[string][]*alignmentStep{}
	if len(workflowIDs) > 0 {
		var workflows []app.Workflow
		if err := s.db.WithContext(ctx).Select("id", "type", "status", "created_at").
			Where(app.Workflow{OwnerID: st.install.ID, OwnerType: "installs", OrgID: st.install.OrgID}).
			Where("id IN ?", workflowIDs).Find(&workflows).Error; err != nil {
			return fmt.Errorf("unable to get explanation workflows: %w", err)
		}
		for i := range workflows {
			workflowsByID[workflows[i].ID] = &workflows[i]
		}
		var steps []alignmentStep
		if err := s.db.WithContext(ctx).Table("install_workflow_steps AS s").
			Select("s.id", "s.install_workflow_id", "s.status", "s.step_target_type", "s.step_target_id", "s.metadata", "s.retried",
				"s.queue_signal->>'type' AS signal_type", "s.queue_signal->'data'->>'ComponentID' AS signal_component_id",
				"COALESCE(NULLIF(s.queue_signal->'data'->>'BuildID', ''), d.component_build_id) AS signal_build_id",
				"d.install_component_id AS resource_id").
			Joins("LEFT JOIN install_deploys d ON d.id = s.step_target_id AND d.deleted_at = 0").
			Where("s.install_workflow_id IN ? AND s.deleted_at = 0", workflowIDs).
			Order("s.group_idx DESC, s.group_retry_idx DESC, s.idx ASC, s.id ASC").Scan(&steps).Error; err != nil {
			return fmt.Errorf("unable to get explanation steps: %w", err)
		}
		for i := range steps {
			stepsByWorkflow[steps[i].InstallWorkflowID] = append(stepsByWorkflow[steps[i].InstallWorkflowID], &steps[i])
		}
	}
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.Category != InstallAlignmentCategoryChangesNotApplied {
			continue
		}
		wf := workflowsByID[byResource[res.ID]]
		if res.OpenDeployment == nil && res.LastAttempt != nil {
			if last := workflowsByID[res.LastAttempt.WorkflowID]; last != nil && (wf == nil || last.CreatedAt.After(wf.CreatedAt)) {
				wf = last
			}
		}
		var steps []*alignmentStep
		if wf != nil {
			steps = stepsByWorkflow[wf.ID]
		}
		res.Explanation = explainUnappliedResource(res, wf, steps)
	}
	return nil
}

func explainUnappliedResource(res *InstallAlignmentResource, wf *app.Workflow, steps []*alignmentStep) *InstallAlignmentExplanation {
	explanation := &InstallAlignmentExplanation{Summary: "We couldn’t determine why these changes haven’t taken effect."}
	if wf == nil {
		if res.LastAttempt == nil && res.AppliedBuildID == "" && res.AppliedConfigID == "" {
			explanation.Summary = "No deployment has started for these changes."
		}
		return explanation
	}
	if res.OpenDeployment == nil && res.appliedAt != nil && wf.CreatedAt.Before(*res.appliedAt) &&
		(res.LastAttempt == nil || !res.LastAttempt.CreatedAt.After(*res.appliedAt)) {
		explanation.Summary = "No deployment has started for these changes."
		return explanation
	}
	var related []*alignmentStep
	for _, step := range steps {
		if step.Retried {
			continue
		}
		matches := false
		switch res.Kind {
		case InstallAlignmentResourceKindSandbox:
			_, matches = alignmentSandboxSignalTypes[step.SignalType]
			matches = matches || step.StepTargetType == "install_sandbox_runs" || step.StepTargetType == "install_sandbox_run"
		case InstallAlignmentResourceKindStack:
			_, matches = alignmentStackSignalTypes[step.SignalType]
			matches = matches || step.StepTargetType == "install_stack_versions" || step.StepTargetType == "install_stack_version"
		default:
			matches = step.ResourceID == res.ID
			for _, t := range alignmentComponentSignalTypes {
				if step.SignalType == t && (step.SignalComponentID == res.ComponentID || (step.Metadata["component_name"] != nil && *step.Metadata["component_name"] == res.Name)) {
					matches = true
				}
			}
		}
		if matches {
			if step.SignalBuildID != "" && res.ExpectedBuildID != "" && step.SignalBuildID != res.ExpectedBuildID {
				return explanation
			}
			related = append(related, step)
		}
	}
	explanation.WorkflowID = wf.ID
	cancelledSummary := "The deployment was cancelled before these changes took effect."
	switch wf.Type {
	case app.WorkflowTypeInputUpdate:
		cancelledSummary = "Applying the input changes was cancelled. The saved values haven’t taken effect."
	case app.WorkflowTypeComponentEnabled:
		cancelledSummary = "Enabling this component was cancelled. It hasn’t been deployed."
	case app.WorkflowTypeComponentDisabled:
		cancelledSummary = "Disabling this component was cancelled. It is still deployed."
	}
	if wf.Status.Status == app.StatusCancelled {
		explanation.Summary = cancelledSummary
		explanation.Status = string(app.StatusCancelled)
		return explanation
	}
	for _, step := range related {
		var summary string
		status := step.Status.Status
		switch step.Status.Status {
		case app.AwaitingApproval:
			summary = "These changes are waiting for approval."
		case app.WorkflowStepApprovalStatusApprovalDenied:
			summary = "These changes weren’t applied because approval was declined."
		case app.StatusUserSkipped:
			summary = "Deployment of this resource was skipped."
		case app.StatusError:
			summary = "The latest deployment failed."
			if workflowAwaitingRetry(wf.Status.Status) {
				summary = "The deployment failed and is waiting to be retried."
				status = app.StatusFailedPendingRetry
			}
		case app.StatusCancelled:
			summary = cancelledSummary
		case app.StatusNotAttempted, app.StatusDiscarded:
			if reason, ok := step.Status.Metadata["reason"].(string); ok && strings.Contains(reason, "workflow stopped before") {
				summary = "The deployment stopped before it reached this resource."
			}
		case app.StatusInProgress, app.StatusPlanning, app.StatusApplying, app.StatusCheckPlan, app.StatusRetrying:
			summary = "The deployment is in progress. These changes haven’t taken effect yet."
		}
		if summary != "" {
			explanation.Summary = summary
			explanation.Status = string(status)
			explanation.StepID = step.ID
			return explanation
		}
	}
	if res.OpenDeployment != nil {
		explanation.Summary = "The deployment is queued and hasn’t started yet."
		explanation.Status = string(app.StatusQueued)
	} else if len(related) == 0 && wf.Status.Status == app.StatusError {
		explanation.Summary = "An earlier operation failed, so deployment of this resource couldn’t start."
		explanation.Status = string(app.StatusError)
	}
	return explanation
}
