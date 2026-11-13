package alignment

import (
	"context"
	"fmt"

	"gorm.io/gorm"

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

	stepsByWorkflow := map[string][]*alignmentStep{}
	deployComponent := map[string]string{}
	if len(openIDs) > 0 {
		steps, err := s.loadAttemptSteps(ctx, func(db *gorm.DB) *gorm.DB { return db.Where("install_workflow_id IN ?", openIDs) })
		if err != nil {
			return err
		}
		for _, step := range steps {
			stepsByWorkflow[step.InstallWorkflowID] = append(stepsByWorkflow[step.InstallWorkflowID], step)
		}
		var deploys []app.InstallDeploy
		if err := s.db.WithContext(ctx).Select("id", "install_component_id").
			Where("install_workflow_id IN ?", openIDs).Find(&deploys).Error; err != nil {
			return fmt.Errorf("unable to get open workflow deploys: %w", err)
		}
		for i := range deploys {
			deployComponent[deploys[i].ID] = deploys[i].InstallComponentID
		}
	}

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.Category != InstallAlignmentCategoryChangesNotApplied {
			continue
		}
		switch {
		case res.OpenDeployment != nil:
			res.Explanation = st.explainOpenDeployment(res, stepsByWorkflow[res.OpenDeployment.ID], deployComponent)
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

func (st *installAlignmentState) explainOpenDeployment(res *InstallAlignmentResource, steps []*alignmentStep, deployComponent map[string]string) *InstallAlignmentExplanation {
	open := res.OpenDeployment
	wf := &app.Workflow{ID: open.ID, Type: open.Type, Status: app.CompositeStatus{Status: app.Status(open.Status)}}
	var related []*alignmentStep
	if target, ok := st.alignmentTargetFor(res); ok {
		for _, step := range steps {
			if stepRelatesToTarget(step, target, deployComponent) {
				related = append(related, step)
			}
		}
	}
	status := string(app.StatusQueued)
	var stepID string
	afterDeny := false
	if len(related) > 0 {
		if outcome, ok := classifyAttempt(wf, related, "", false); ok && outcome.status != attemptStatusUnknown {
			status, stepID, afterDeny = outcome.status, outcome.stepID, outcome.afterDeny
		} else {
			status = string(app.StatusInProgress)
		}
	}
	return &InstallAlignmentExplanation{
		Summary:    attemptSummary(res, status, open.Type, afterDeny),
		Status:     status,
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
