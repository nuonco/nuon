package alignment

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

const (
	maxAttemptsPerResource = 25
	attemptStatusUnknown   = "unknown"
)

type attemptCandidate struct {
	targetID     string
	workflowID   string
	createdAt    time.Time
	buildID      string
	status       string
	policySkip   bool
	workflowOnly bool
}

type attemptOutcome struct {
	status    string
	stepID    string
	afterDeny bool
}

var attemptFailureStatuses = map[string]struct{}{
	string(app.StatusError):                               {},
	string(app.StatusCancelled):                           {},
	string(app.WorkflowStepApprovalStatusApprovalDenied):  {},
	string(app.WorkflowStepApprovalStatusApprovalExpired): {},
	string(app.InstallDeployStatusHealthFailed):           {},
}

var attemptActiveStepStatuses = map[app.Status]struct{}{
	"":                   {},
	app.StatusPending:    {},
	app.StatusQueued:     {},
	app.StatusInProgress: {},
	app.StatusRetrying:   {},
	app.StatusPlanning:   {},
	app.StatusApplying:   {},
	app.StatusCheckPlan:  {},
	app.AwaitingApproval: {},
}

func deployTypeWanted(res *InstallAlignmentResource, typ app.InstallDeployType, status string) bool {
	switch app.InstallDeployStatus(status) {
	case app.InstallDeployStatusRetried, app.InstallDeployStatusDriftDetected, app.InstallDeployStatusNoDrift,
		app.InstallDeployStatus(app.WorkflowStepDrifted):
		return false
	}
	if res.wantsTeardown {
		return typ == app.InstallDeployTypeTeardown
	}
	return typ != app.InstallDeployTypeTeardown && typ != app.InstallDeployTypeRecover && res.ExpectedBuildID != ""
}

func (s *Helpers) equivalentBuilds(ctx context.Context, st *installAlignmentState, buildIDs map[string]struct{}) (map[string]bool, error) {
	out := map[string]bool{}
	if len(buildIDs) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(buildIDs))
	for id := range buildIDs {
		ids = append(ids, id)
	}
	var builds []app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Select("id", "component_config_connection_id", "source_checksum", "source_digest", "source_identity").
		Preload("ComponentConfigConnection", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped().Select("id", "component_id", "checksum")
		}).
		Where("id IN ?", ids).Find(&builds).Error; err != nil {
		return nil, fmt.Errorf("unable to get attempt builds: %w", err)
	}
	for i := range builds {
		b := &builds[i]
		target := st.targets[b.ComponentConfigConnection.ComponentID]
		if target == nil || target.Build == nil || target.Config == nil || target.Config.Checksum == "" {
			continue
		}
		out[b.ID] = b.ComponentConfigConnection.Checksum == target.Config.Checksum && configdiff.ComponentSourcesEqual(b, target.Build)
	}
	return out, nil
}

func (s *Helpers) loadBlockingAttempts(ctx context.Context, st *installAlignmentState, alignment *InstallAlignment, inputAttempts map[string]*app.Workflow) error {
	byID := map[string]*InstallAlignmentResource{}
	var componentIDs []string
	var minSince *time.Time
	unbounded := false
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.Category != InstallAlignmentCategoryChangesNotApplied {
			continue
		}
		byID[res.ID] = res
		if res.Kind == InstallAlignmentResourceKindComponent || res.Kind == InstallAlignmentResourceKindImage {
			componentIDs = append(componentIDs, res.ID)
			switch {
			case res.attemptsSince == nil:
				unbounded = true
			case minSince == nil || res.attemptsSince.Before(*minSince):
				minSince = res.attemptsSince
			}
		}
	}
	if len(byID) == 0 {
		return nil
	}

	if unbounded {
		minSince = nil
	}
	candidates := map[string][]*attemptCandidate{}
	if len(componentIDs) > 0 {
		var rows []struct {
			ID                 string
			InstallComponentID string
			ComponentBuildID   string
			Type               app.InstallDeployType
			Status             string
			PolicyAction       string
			CreatedAt          time.Time
			InstallWorkflowID  *string
		}
		query := s.db.WithContext(ctx).Table("install_deploys AS d").
			Select("d.id", "d.install_component_id", "d.component_build_id", "d.type",
				"COALESCE(NULLIF(d.status_v2->>'status', ''), d.status) AS status",
				"COALESCE(d.status_v2->'metadata'->>'policy_action', '') AS policy_action",
				"d.created_at", "d.install_workflow_id").
			Joins("JOIN install_workflows w ON w.id = d.install_workflow_id").
			Where("w.owner_id = ? AND w.owner_type = ? AND w.deleted_at = 0 AND NOT w.plan_only", st.install.ID, "installs").
			Where("d.deleted_at = 0 AND d.install_component_id IN ?", componentIDs)
		if minSince != nil {
			query = query.Where("d.created_at > ?", *minSince)
		}
		if err := query.Order("d.created_at DESC").Scan(&rows).Error; err != nil {
			return fmt.Errorf("unable to get install deploy attempts: %w", err)
		}
		otherBuilds := map[string]struct{}{}
		for _, row := range rows {
			if res := byID[row.InstallComponentID]; res != nil && !res.wantsTeardown && row.ComponentBuildID != res.ExpectedBuildID {
				otherBuilds[row.ComponentBuildID] = struct{}{}
			}
		}
		equivalent, err := s.equivalentBuilds(ctx, st, otherBuilds)
		if err != nil {
			return err
		}
		for _, row := range rows {
			res := byID[row.InstallComponentID]
			if res == nil || len(candidates[res.ID]) >= maxAttemptsPerResource || !deployTypeWanted(res, row.Type, row.Status) ||
				(res.attemptsSince != nil && !row.CreatedAt.After(*res.attemptsSince)) {
				continue
			}
			if !res.wantsTeardown && row.ComponentBuildID != res.ExpectedBuildID && !equivalent[row.ComponentBuildID] {
				continue
			}
			c := &attemptCandidate{
				targetID:   row.ID,
				createdAt:  row.CreatedAt,
				buildID:    row.ComponentBuildID,
				status:     row.Status,
				policySkip: row.PolicyAction == "skip_apply",
			}
			if row.InstallWorkflowID != nil {
				c.workflowID = *row.InstallWorkflowID
			}
			candidates[res.ID] = append(candidates[res.ID], c)
		}
	}

	if st.sandbox != nil && byID[st.sandbox.ID] != nil {
		res := byID[st.sandbox.ID]
		query := s.db.WithContext(ctx).
			Select("install_sandbox_runs.id", "install_sandbox_runs.created_at", "install_sandbox_runs.install_workflow_id",
				"install_sandbox_runs.status", "install_sandbox_runs.status_v2").
			Joins("JOIN install_workflows w ON w.id = install_sandbox_runs.install_workflow_id AND w.deleted_at = 0 AND NOT w.plan_only").
			Where(app.InstallSandboxRun{InstallSandboxID: &st.sandbox.ID}).
			Where("install_sandbox_runs.run_type IS DISTINCT FROM ?", app.SandboxRunTypeDeprovision)
		if res.attemptsSince != nil {
			query = query.Where("install_sandbox_runs.created_at > ?", *res.attemptsSince)
		}
		if st.expectedSandbox != nil {
			query = query.Where(app.InstallSandboxRun{AppSandboxConfigID: st.expectedSandbox.ID})
		}
		var runs []app.InstallSandboxRun
		if err := query.Order("install_sandbox_runs.created_at DESC").Limit(maxAttemptsPerResource).Find(&runs).Error; err != nil {
			return fmt.Errorf("unable to get install sandbox run attempts: %w", err)
		}
		for i := range runs {
			c := &attemptCandidate{
				targetID:   runs[i].ID,
				createdAt:  runs[i].CreatedAt,
				status:     string(runs[i].Status),
				policySkip: runs[i].StatusV2.Metadata["policy_action"] == "skip_apply",
			}
			if runs[i].InstallWorkflowID != nil {
				c.workflowID = *runs[i].InstallWorkflowID
			}
			candidates[res.ID] = append(candidates[res.ID], c)
		}
	}

	if st.stack != nil && byID[st.stack.ID] != nil && st.install.AppConfigID != "" {
		res := byID[st.stack.ID]
		query := s.db.WithContext(ctx).
			Select("id", "created_at", "status").
			Where(app.InstallStackVersion{InstallStackID: st.stack.ID, AppConfigID: st.install.AppConfigID})
		if res.attemptsSince != nil {
			query = query.Where("created_at > ?", *res.attemptsSince)
		}
		var versions []app.InstallStackVersion
		if err := query.Order("created_at DESC").Limit(maxAttemptsPerResource).Find(&versions).Error; err != nil {
			return fmt.Errorf("unable to get install stack version attempts: %w", err)
		}
		for i := range versions {
			candidates[res.ID] = append(candidates[res.ID], &attemptCandidate{
				targetID:  versions[i].ID,
				createdAt: versions[i].CreatedAt,
				status:    string(versions[i].Status.Status),
			})
		}
	}

	for id, wf := range inputAttempts {
		res := byID[id]
		if res == nil || (res.attemptsSince != nil && !wf.CreatedAt.After(*res.attemptsSince)) {
			continue
		}
		candidates[id] = append(candidates[id], &attemptCandidate{
			targetID:     wf.ID,
			workflowID:   wf.ID,
			createdAt:    wf.CreatedAt,
			status:       string(wf.Status.Status),
			workflowOnly: true,
		})
	}

	var targetIDs, workflowOnlyIDs []string
	for _, list := range candidates {
		for _, c := range list {
			if c.workflowOnly {
				workflowOnlyIDs = append(workflowOnlyIDs, c.workflowID)
			} else {
				targetIDs = append(targetIDs, c.targetID)
			}
		}
	}
	stepsByTarget := map[string][]*alignmentStep{}
	stepsByWorkflow := map[string][]*alignmentStep{}
	deployComponent := map[string]string{}
	if len(targetIDs) > 0 {
		steps, err := s.loadAttemptSteps(ctx, func(db *gorm.DB) *gorm.DB { return db.Where("step_target_id IN ?", targetIDs) })
		if err != nil {
			return err
		}
		for _, step := range steps {
			stepsByTarget[step.StepTargetID] = append(stepsByTarget[step.StepTargetID], step)
		}
	}
	if len(workflowOnlyIDs) > 0 {
		steps, err := s.loadAttemptSteps(ctx, func(db *gorm.DB) *gorm.DB { return db.Where("install_workflow_id IN ?", workflowOnlyIDs) })
		if err != nil {
			return err
		}
		for _, step := range steps {
			stepsByWorkflow[step.InstallWorkflowID] = append(stepsByWorkflow[step.InstallWorkflowID], step)
		}
		var deploys []app.InstallDeploy
		if err := s.db.WithContext(ctx).Select("id", "install_component_id").
			Where("install_workflow_id IN ?", workflowOnlyIDs).Find(&deploys).Error; err != nil {
			return fmt.Errorf("unable to get input workflow deploys: %w", err)
		}
		for i := range deploys {
			deployComponent[deploys[i].ID] = deploys[i].InstallComponentID
		}
	}

	workflowIDs := map[string]struct{}{}
	for _, list := range candidates {
		for _, c := range list {
			if c.workflowID == "" {
				for _, step := range stepsByTarget[c.targetID] {
					if step.InstallWorkflowID != "" {
						c.workflowID = step.InstallWorkflowID
						break
					}
				}
			}
			if c.workflowID != "" {
				workflowIDs[c.workflowID] = struct{}{}
			}
		}
	}
	workflowsByID := map[string]*app.Workflow{}
	if len(workflowIDs) > 0 {
		ids := make([]string, 0, len(workflowIDs))
		for id := range workflowIDs {
			ids = append(ids, id)
		}
		var workflows []app.Workflow
		if err := s.db.WithContext(ctx).Select("id", "type", "status", "created_at").
			Where("id IN ?", ids).Find(&workflows).Error; err != nil {
			return fmt.Errorf("unable to get attempt workflows: %w", err)
		}
		for i := range workflows {
			workflowsByID[workflows[i].ID] = &workflows[i]
		}
	}

	for id, list := range candidates {
		res := byID[id]
		target, ok := st.alignmentTargetFor(res)
		if !ok {
			continue
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].createdAt.After(list[j].createdAt) })
		deployWorkflows := map[string]struct{}{}
		for _, c := range list {
			if !c.workflowOnly && c.workflowID != "" {
				deployWorkflows[c.workflowID] = struct{}{}
			}
		}
		for _, c := range list {
			if res.OpenDeployment != nil && c.workflowID == res.OpenDeployment.ID {
				continue
			}
			var steps []*alignmentStep
			if c.workflowOnly {
				if _, ok := deployWorkflows[c.workflowID]; ok {
					continue
				}
				for _, step := range stepsByWorkflow[c.workflowID] {
					if stepRelatesToTarget(step, target, deployComponent) {
						steps = append(steps, step)
					}
				}
			} else {
				steps = stepsByTarget[c.targetID]
			}
			wf := workflowsByID[c.workflowID]
			fallback := c.status
			if c.workflowOnly {
				fallback = ""
			}
			outcome, ok := classifyAttempt(wf, steps, fallback, c.policySkip)
			if !ok {
				continue
			}
			if res.BlockingAttempt != nil {
				res.EarlierAttempts++
				continue
			}
			res.BlockingAttempt = &InstallAlignmentAttempt{
				ID:         c.targetID,
				Status:     outcome.status,
				CreatedAt:  c.createdAt,
				WorkflowID: c.workflowID,
				StepID:     outcome.stepID,
				BuildID:    c.buildID,
			}
			if wf != nil {
				res.BlockingAttempt.Type = wf.Type
			}
			res.blockingAfterDeny = outcome.afterDeny
		}
	}
	return nil
}

func (s *Helpers) loadAttemptSteps(ctx context.Context, scope func(*gorm.DB) *gorm.DB) ([]*alignmentStep, error) {
	var steps []alignmentStep
	if err := scope(s.db.WithContext(ctx).Model(&app.WorkflowStep{})).
		Select("id", "install_workflow_id", "status", "step_target_type", "step_target_id", "metadata", "retried",
			"queue_signal->>'type' AS signal_type",
			"queue_signal->'data'->>'ComponentID' AS signal_component_id").
		Scan(&steps).Error; err != nil {
		return nil, fmt.Errorf("unable to get attempt steps: %w", err)
	}
	out := make([]*alignmentStep, len(steps))
	for i := range steps {
		out[i] = &steps[i]
	}
	return out, nil
}

func (st *installAlignmentState) alignmentTargetFor(res *InstallAlignmentResource) (alignmentTarget, bool) {
	switch res.Kind {
	case InstallAlignmentResourceKindStack:
		return alignmentTarget{kind: res.Kind, stack: st.stack}, st.stack != nil
	case InstallAlignmentResourceKindSandbox:
		return alignmentTarget{kind: res.Kind, sandbox: st.sandbox}, st.sandbox != nil
	default:
		for i := range st.components {
			if st.components[i].ID == res.ID {
				return alignmentTarget{kind: res.Kind, installComponent: &st.components[i]}, true
			}
		}
		return alignmentTarget{}, false
	}
}

func stepDenied(step *alignmentStep) bool {
	denied, ok := step.Status.Metadata["deny_violations"].([]any)
	return ok && len(denied) > 0
}

func classifyAttempt(wf *app.Workflow, steps []*alignmentStep, fallback string, policySkip bool) (attemptOutcome, bool) {
	var live []*alignmentStep
	for _, step := range steps {
		if !step.Retried {
			live = append(live, step)
		}
	}
	if len(steps) > 0 && len(live) == 0 {
		return attemptOutcome{}, false
	}
	var wfStatus app.Status
	if wf != nil {
		wfStatus = wf.Status.Status
	}
	find := func(statuses ...app.Status) *alignmentStep {
		for _, step := range live {
			for _, status := range statuses {
				if step.Status.Status == status {
					return step
				}
			}
		}
		return nil
	}

	if step := find(app.StatusError); step != nil {
		switch {
		case stepDenied(step):
			return attemptOutcome{status: "policy-failed", stepID: step.ID}, true
		case workflowAwaitingRetry(wfStatus):
			return attemptOutcome{status: string(app.StatusFailedPendingRetry), stepID: step.ID}, true
		default:
			return attemptOutcome{status: string(app.StatusError), stepID: step.ID}, true
		}
	}
	for _, status := range []app.Status{app.WorkflowStepApprovalStatusApprovalDenied, app.WorkflowStepApprovalStatusApprovalExpired} {
		if step := find(status); step != nil {
			return attemptOutcome{status: string(status), stepID: step.ID}, true
		}
	}
	if step := find(app.StatusUserSkipped); step != nil {
		return attemptOutcome{status: string(app.StatusUserSkipped), stepID: step.ID, afterDeny: stepDenied(step)}, true
	}
	if step := find(app.StatusCancelled); step != nil {
		return attemptOutcome{status: string(app.StatusCancelled), stepID: step.ID}, true
	}
	if step := find(app.StatusAutoSkipped); step != nil {
		if policySkip || step.Status.Metadata["check"] == "policy-skip-apply" || step.Status.Metadata["policy_action"] == "skip_apply" {
			return attemptOutcome{status: "policy-skipped", stepID: step.ID}, true
		}
		return attemptOutcome{status: string(app.StatusAutoSkipped), stepID: step.ID}, true
	}
	if step := find(app.StatusNotAttempted, app.StatusDiscarded); step != nil {
		return attemptOutcome{status: string(app.StatusNotAttempted), stepID: step.ID}, true
	}
	if _, failed := attemptFailureStatuses[fallback]; failed {
		return attemptOutcome{status: fallback}, true
	}
	succeeded := find(app.StatusSuccess) != nil
	for _, step := range live {
		_, active := attemptActiveStepStatuses[step.Status.Status]
		if !active && (succeeded || step.Status.Status != app.WorkflowStepApprovalStatusApproved) {
			continue
		}
		switch {
		case wfStatus == app.StatusCancelled:
			return attemptOutcome{status: string(app.StatusCancelled), stepID: step.ID}, true
		case wfStatus == app.StatusError || wfStatus == app.StatusSuccess:
			return attemptOutcome{status: string(app.StatusNotAttempted), stepID: step.ID}, true
		case step.Status.Status == app.AwaitingApproval:
			return attemptOutcome{status: string(app.AwaitingApproval), stepID: step.ID}, true
		default:
			return attemptOutcome{status: string(app.StatusInProgress), stepID: step.ID}, true
		}
	}

	if wf == nil {
		return attemptOutcome{}, false
	}
	switch {
	case wfStatus == app.StatusCancelled:
		return attemptOutcome{status: string(app.StatusCancelled)}, true
	case len(live) > 0:
		return attemptOutcome{status: attemptStatusUnknown}, true
	case wfStatus == app.StatusError || workflowAwaitingRetry(wfStatus):
		return attemptOutcome{status: string(app.StatusNotAttempted)}, true
	case wfStatus == app.StatusSuccess:
		return attemptOutcome{}, false
	default:
		return attemptOutcome{status: string(app.StatusQueued)}, true
	}
}

func (s *Helpers) loadBuildSummaries(ctx context.Context, alignment *InstallAlignment) error {
	ids := []string{}
	for _, res := range alignment.Resources {
		for _, id := range []string{res.ExpectedBuildID, res.AppliedBuildID} {
			if id != "" {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var builds []app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Select("id", "created_at", "vcs_connection_commit_id", "resolved_tag").
		Preload("VCSConnectionCommit", func(db *gorm.DB) *gorm.DB { return db.Select("id", "sha") }).
		Where("id IN ?", ids).
		Find(&builds).Error; err != nil {
		return fmt.Errorf("unable to get alignment builds: %w", err)
	}
	summaries := make(map[string]*InstallAlignmentBuild, len(builds))
	for i := range builds {
		b := &builds[i]
		summary := &InstallAlignmentBuild{ID: b.ID, ResolvedTag: b.ResolvedTag, CreatedAt: b.CreatedAt}
		if b.VCSConnectionCommit != nil {
			summary.SHA = b.VCSConnectionCommit.SHA
		}
		summaries[b.ID] = summary
	}
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		res.ExpectedBuild = summaries[res.ExpectedBuildID]
		res.AppliedBuild = summaries[res.AppliedBuildID]
	}
	return nil
}
