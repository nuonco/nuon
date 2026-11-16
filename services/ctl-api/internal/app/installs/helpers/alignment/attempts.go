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
	denied       bool
	applied      bool
	workflowOnly bool
}

type attemptOutcome struct {
	status    string
	afterDeny bool
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
			Denied             bool
			Applied            bool
			CreatedAt          time.Time
			InstallWorkflowID  *string
		}
		query := s.db.WithContext(ctx).Table("install_deploys AS d").
			Select("d.id", "d.install_component_id", "d.component_build_id", "d.type",
				"COALESCE(NULLIF(d.status_v2->>'status', ''), d.status) AS status",
				"COALESCE(d.status_v2->'metadata'->>'policy_action', '') AS policy_action",
				"CASE WHEN jsonb_typeof(d.status_v2->'metadata'->'deny_violations') = 'array' THEN jsonb_array_length(d.status_v2->'metadata'->'deny_violations') > 0 ELSE false END AS denied",
				"d.applied_at IS NOT NULL AS applied",
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
				denied:     row.Denied,
				applied:    row.Applied,
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
				"install_sandbox_runs.status", "install_sandbox_runs.status_v2", "install_sandbox_runs.applied_at").
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
			status := string(runs[i].StatusV2.Status)
			if status == "" {
				status = string(runs[i].Status)
			}
			c := &attemptCandidate{
				targetID:   runs[i].ID,
				createdAt:  runs[i].CreatedAt,
				status:     status,
				policySkip: runs[i].StatusV2.Metadata["policy_action"] == "skip_apply",
				denied:     metadataDenied(runs[i].StatusV2.Metadata),
				applied:    runs[i].AppliedAt != nil,
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
				targetID:   versions[i].ID,
				createdAt:  versions[i].CreatedAt,
				status:     string(versions[i].Status.Status),
				policySkip: versions[i].Status.Metadata["policy_action"] == "skip_apply",
				denied:     metadataDenied(versions[i].Status.Metadata),
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
			workflowOnly: true,
		})
	}

	var targetIDs []string
	for _, list := range candidates {
		for _, c := range list {
			if !c.workflowOnly {
				targetIDs = append(targetIDs, c.targetID)
			}
		}
	}
	links, err := s.attemptStepLinks(ctx, targetIDs)
	if err != nil {
		return err
	}

	workflowIDs := map[string]struct{}{}
	for _, list := range candidates {
		for _, c := range list {
			if c.workflowID == "" {
				c.workflowID = links[c.targetID].workflowID
			}
			if c.workflowID != "" {
				workflowIDs[c.workflowID] = struct{}{}
			}
		}
	}
	workflowsByID, err := s.attemptWorkflows(ctx, workflowIDs)
	if err != nil {
		return err
	}

	for id, list := range candidates {
		res := byID[id]
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
			wf := workflowsByID[c.workflowID]
			var outcome attemptOutcome
			ok := false
			if c.workflowOnly {
				if _, ok := deployWorkflows[c.workflowID]; ok {
					continue
				}
				outcome, ok = workflowOutcome(wf)
			} else {
				outcome, ok = rowOutcome(c, wf)
			}
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
				StepID:     links[c.targetID].stepID,
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

type attemptStepLink struct {
	stepID     string
	workflowID string
}

func (s *Helpers) attemptStepLinks(ctx context.Context, targetIDs []string) (map[string]attemptStepLink, error) {
	links := map[string]attemptStepLink{}
	if len(targetIDs) == 0 {
		return links, nil
	}
	var steps []struct {
		ID                string
		InstallWorkflowID string
		StepTargetID      string
	}
	if err := s.db.WithContext(ctx).Model(&app.WorkflowStep{}).
		Select("id", "install_workflow_id", "step_target_id").
		Where("step_target_id IN ? AND NOT retried", targetIDs).
		Order("idx DESC, created_at DESC").
		Scan(&steps).Error; err != nil {
		return nil, fmt.Errorf("unable to get attempt steps: %w", err)
	}
	for _, step := range steps {
		if _, ok := links[step.StepTargetID]; !ok {
			links[step.StepTargetID] = attemptStepLink{stepID: step.ID, workflowID: step.InstallWorkflowID}
		}
	}
	return links, nil
}

func (s *Helpers) attemptWorkflows(ctx context.Context, ids map[string]struct{}) (map[string]*app.Workflow, error) {
	out := map[string]*app.Workflow{}
	if len(ids) == 0 {
		return out, nil
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	var workflows []app.Workflow
	if err := s.db.WithContext(ctx).Select("id", "type", "status", "created_at").
		Where("id IN ?", list).Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("unable to get attempt workflows: %w", err)
	}
	for i := range workflows {
		out[workflows[i].ID] = &workflows[i]
	}
	return out, nil
}

func metadataDenied(meta map[string]any) bool {
	denied, ok := meta["deny_violations"].([]any)
	return ok && len(denied) > 0
}

func rowOutcome(c *attemptCandidate, wf *app.Workflow) (attemptOutcome, bool) {
	if c.applied {
		return attemptOutcome{}, false
	}
	var wfStatus app.Status
	if wf != nil {
		wfStatus = wf.Status.Status
	}
	switch app.Status(c.status) {
	case app.StatusError, app.Status(app.InstallDeployStatusHealthFailed), app.Status(app.SandboxRunStatusAccessError):
		switch {
		case c.denied:
			return attemptOutcome{status: "policy-failed"}, true
		case workflowAwaitingRetry(wfStatus):
			return attemptOutcome{status: string(app.StatusFailedPendingRetry)}, true
		}
		return attemptOutcome{status: string(app.StatusError)}, true
	case app.WorkflowStepApprovalStatusApprovalDenied, app.WorkflowStepApprovalStatusApprovalExpired, app.StatusCancelled:
		return attemptOutcome{status: c.status}, true
	case app.StatusUserSkipped:
		return attemptOutcome{status: c.status, afterDeny: c.denied}, true
	case app.StatusAutoSkipped:
		if c.policySkip {
			return attemptOutcome{status: "policy-skipped"}, true
		}
		return attemptOutcome{status: c.status}, true
	case app.StatusNotAttempted, app.StatusDiscarded, app.InstallStackVersionStatusOutdated, app.InstallStackVersionStatusExpired:
		return attemptOutcome{status: string(app.StatusNotAttempted)}, true
	case app.Status(app.InstallDeployStatusActive), app.Status(app.InstallDeployStatusInactive), app.Status(app.InstallDeployStatusNoop),
		app.StatusSuccess, app.Status(app.SandboxRunStatusDeprovisioned):
		return attemptOutcome{}, false
	case "", app.StatusQueued, app.StatusPending, app.Status(app.InstallDeployStatusPendingApproval), app.AwaitingApproval,
		app.StatusPlanning, app.StatusApplying, app.StatusInProgress, app.StatusRetrying, app.StatusCheckPlan,
		app.Status(app.InstallDeployStatusSyncing), app.Status(app.InstallDeployStatusExecuting),
		app.Status(app.SandboxRunStatusProvisioning), app.Status(app.SandboxRunStatusReprovisioning), app.Status(app.SandboxRunStatusDeprovisioning),
		app.InstallStackVersionStatusGenerating, app.InstallStackVersionStatusPendingUser:
		return unfinishedOutcome(c.status, wfStatus), true
	}
	return attemptOutcome{status: attemptStatusUnknown}, true
}

func unfinishedOutcome(status string, wfStatus app.Status) attemptOutcome {
	switch {
	case wfStatus == app.StatusCancelled:
		return attemptOutcome{status: string(app.StatusCancelled)}
	case wfStatus == "" || wfStatus == app.StatusError || wfStatus == app.StatusSuccess || workflowAwaitingRetry(wfStatus):
		return attemptOutcome{status: string(app.StatusNotAttempted)}
	case status == string(app.InstallDeployStatusPendingApproval) || status == string(app.AwaitingApproval) || wfStatus == app.AwaitingApproval:
		return attemptOutcome{status: string(app.AwaitingApproval)}
	case status == "" || status == string(app.StatusQueued) || status == string(app.StatusPending):
		return attemptOutcome{status: string(app.StatusQueued)}
	default:
		return attemptOutcome{status: string(app.StatusInProgress)}
	}
}

func workflowOutcome(wf *app.Workflow) (attemptOutcome, bool) {
	if wf == nil {
		return attemptOutcome{}, false
	}
	switch status := wf.Status.Status; {
	case status == app.StatusSuccess:
		return attemptOutcome{}, false
	case status == app.StatusCancelled, status == app.StatusError,
		status == app.WorkflowStepApprovalStatusApprovalDenied, status == app.WorkflowStepApprovalStatusApprovalExpired:
		return attemptOutcome{status: string(status)}, true
	case workflowAwaitingRetry(status):
		return attemptOutcome{status: string(app.StatusFailedPendingRetry)}, true
	case status == app.AwaitingApproval:
		return attemptOutcome{status: string(app.AwaitingApproval)}, true
	case status == "" || status == app.StatusPending || status == app.StatusQueued:
		return attemptOutcome{status: string(app.StatusQueued)}, true
	default:
		return attemptOutcome{status: string(app.StatusInProgress)}, true
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
