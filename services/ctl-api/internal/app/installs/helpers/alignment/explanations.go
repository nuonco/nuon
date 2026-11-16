package alignment

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) explainInstallAlignment(ctx context.Context, st *installAlignmentState, alignment *InstallAlignment) error {
	var openIDs []string
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		switch res.Category {
		case InstallAlignmentCategoryChangesNotApplied:
			if res.OpenDeployment != nil {
				openIDs = append(openIDs, res.OpenDeployment.ID)
			}
		case InstallAlignmentCategoryStateDriftDetected:
			res.Explanation = &InstallAlignmentExplanation{Summary: "A drift check found differences between live resources and what was last applied."}
			if res.Plan != nil && res.Plan.PlanOnly {
				res.Explanation.WorkflowID = res.Plan.WorkflowID
				res.Explanation.StepID = res.Plan.StepID
			}
		}
	}

	openRows, links, err := s.loadOpenDeploymentRows(ctx, st, openIDs)
	if err != nil {
		return err
	}

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.Category != InstallAlignmentCategoryChangesNotApplied {
			continue
		}
		switch {
		case res.OpenDeployment != nil:
			res.Explanation = explainOpenDeployment(res, openRows[res.ID], links)
		case res.BlockingAttempt != nil:
			a := res.BlockingAttempt
			res.Explanation = &InstallAlignmentExplanation{
				Summary:    attemptSummary(res, a.Status, a.Type, res.blockingAfterDeny),
				Status:     a.Status,
				WorkflowID: a.WorkflowID,
				StepID:     a.StepID,
			}
		case res.wantsTeardown:
			res.Explanation = &InstallAlignmentExplanation{Summary: "No teardown has started for this component."}
		default:
			res.Explanation = &InstallAlignmentExplanation{Summary: fmt.Sprintf("No %s has started for these changes.", attemptNoun(res))}
		}
	}
	return nil
}

func (s *Helpers) loadOpenDeploymentRows(ctx context.Context, st *installAlignmentState, openIDs []string) (map[string]*attemptCandidate, map[string]attemptStepLink, error) {
	rows := map[string]*attemptCandidate{}
	if len(openIDs) == 0 {
		return rows, nil, nil
	}
	keep := func(resourceID string, c *attemptCandidate) {
		if prev := rows[resourceID]; prev == nil || c.createdAt.After(prev.createdAt) {
			rows[resourceID] = c
		}
	}
	var deploys []struct {
		ID                 string
		InstallComponentID string
		Status             string
		PolicyAction       string
		Denied             bool
		Applied            bool
		CreatedAt          time.Time
		InstallWorkflowID  string
	}
	if err := s.db.WithContext(ctx).Table("install_deploys AS d").
		Select("d.id", "d.install_component_id",
			"COALESCE(NULLIF(d.status_v2->>'status', ''), d.status) AS status",
			"COALESCE(d.status_v2->'metadata'->>'policy_action', '') AS policy_action",
			"CASE WHEN jsonb_typeof(d.status_v2->'metadata'->'deny_violations') = 'array' THEN jsonb_array_length(d.status_v2->'metadata'->'deny_violations') > 0 ELSE false END AS denied",
			"d.applied_at IS NOT NULL AS applied",
			"d.created_at", "d.install_workflow_id").
		Where("d.deleted_at = 0 AND d.install_workflow_id IN ?", openIDs).
		Where("COALESCE(NULLIF(d.status_v2->>'status', ''), d.status) IS DISTINCT FROM ?", app.InstallDeployStatusRetried).
		Scan(&deploys).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to get open workflow deploys: %w", err)
	}
	var targetIDs []string
	for _, d := range deploys {
		targetIDs = append(targetIDs, d.ID)
		keep(d.InstallComponentID, &attemptCandidate{
			targetID: d.ID, workflowID: d.InstallWorkflowID, createdAt: d.CreatedAt, status: d.Status,
			policySkip: d.PolicyAction == "skip_apply", denied: d.Denied, applied: d.Applied,
		})
	}
	if st.sandbox != nil {
		var runs []app.InstallSandboxRun
		if err := s.db.WithContext(ctx).
			Select("id", "created_at", "install_workflow_id", "status", "status_v2", "applied_at").
			Where(app.InstallSandboxRun{InstallSandboxID: &st.sandbox.ID}).
			Where("install_workflow_id IN ?", openIDs).
			Where("status IS DISTINCT FROM ?", app.SandboxRunStatusRetried).
			Find(&runs).Error; err != nil {
			return nil, nil, fmt.Errorf("unable to get open workflow sandbox runs: %w", err)
		}
		for i := range runs {
			status := string(runs[i].StatusV2.Status)
			if status == "" {
				status = string(runs[i].Status)
			}
			targetIDs = append(targetIDs, runs[i].ID)
			keep(st.sandbox.ID, &attemptCandidate{
				targetID: runs[i].ID, createdAt: runs[i].CreatedAt, status: status,
				policySkip: runs[i].StatusV2.Metadata["policy_action"] == "skip_apply",
				denied:     metadataDenied(runs[i].StatusV2.Metadata),
				applied:    runs[i].AppliedAt != nil,
			})
		}
	}
	links, err := s.attemptStepLinks(ctx, targetIDs)
	if err != nil {
		return nil, nil, err
	}
	return rows, links, nil
}

func explainOpenDeployment(res *InstallAlignmentResource, row *attemptCandidate, links map[string]attemptStepLink) *InstallAlignmentExplanation {
	open := res.OpenDeployment
	wf := &app.Workflow{ID: open.ID, Type: open.Type, Status: app.CompositeStatus{Status: app.Status(open.Status)}}
	outcome := attemptOutcome{status: string(app.StatusQueued)}
	var stepID string
	if row != nil {
		stepID = links[row.targetID].stepID
		if o, ok := rowOutcome(row, wf); ok && o.status != attemptStatusUnknown {
			outcome = o
		} else {
			outcome = attemptOutcome{status: string(app.StatusInProgress)}
		}
	}
	return &InstallAlignmentExplanation{
		Summary:    attemptSummary(res, outcome.status, open.Type, outcome.afterDeny),
		Status:     outcome.status,
		WorkflowID: open.ID,
		StepID:     stepID,
	}
}

func attemptNoun(res *InstallAlignmentResource) string {
	switch {
	case res.wantsTeardown:
		return "teardown"
	case res.Kind == InstallAlignmentResourceKindImage:
		return "image sync"
	default:
		return "deployment"
	}
}

func attemptSummary(res *InstallAlignmentResource, status string, wfType app.WorkflowType, afterDeny bool) string {
	noun := attemptNoun(res)
	switch app.Status(status) {
	case app.StatusError:
		return fmt.Sprintf("The latest %s failed.", noun)
	case "policy-failed":
		return fmt.Sprintf("A policy denied the latest %s.", noun)
	case app.StatusFailedPendingRetry:
		return fmt.Sprintf("The %s failed and is waiting to be retried.", noun)
	case app.StatusCancelled:
		switch wfType {
		case app.WorkflowTypeInputUpdate:
			return "Applying the input changes was cancelled. The saved values haven’t taken effect."
		case app.WorkflowTypeComponentEnabled:
			return "Enabling this component was cancelled. It hasn’t been deployed."
		case app.WorkflowTypeComponentDisabled:
			return "Disabling this component was cancelled. It is still deployed."
		}
		return fmt.Sprintf("The %s was cancelled before these changes took effect.", noun)
	case app.WorkflowStepApprovalStatusApprovalDenied:
		return "These changes weren’t applied because approval was declined."
	case app.WorkflowStepApprovalStatusApprovalExpired:
		return "Approval expired before these changes were applied."
	case app.StatusUserSkipped:
		if afterDeny {
			return fmt.Sprintf("A policy denied the latest %s, and it was skipped.", noun)
		}
		return fmt.Sprintf("The %s of this resource was skipped.", noun)
	case "policy-skipped":
		return "A policy skipped the apply, so the existing deployment was kept."
	case app.StatusAutoSkipped:
		return fmt.Sprintf("The plan had no changes, so the %s was skipped.", noun)
	case app.StatusNotAttempted:
		return fmt.Sprintf("The %s stopped before it reached this resource.", noun)
	case app.AwaitingApproval:
		return "These changes are waiting for approval."
	case app.StatusInProgress:
		return fmt.Sprintf("The %s is in progress. These changes haven’t taken effect yet.", noun)
	case app.StatusQueued:
		return fmt.Sprintf("The %s is queued and hasn’t started yet.", noun)
	default:
		return fmt.Sprintf("We couldn’t determine the outcome of the latest %s.", noun)
	}
}
