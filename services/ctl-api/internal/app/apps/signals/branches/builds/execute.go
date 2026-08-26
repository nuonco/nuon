package builds

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/sandboxbuild"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/queuebuild"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const buildBatchSize = 5

// buildEntry tracks a single component build for metadata updates.
type buildEntry struct {
	BuildID       string  `json:"build_id,omitempty"`
	ComponentID   string  `json:"component_id"`
	ComponentName string  `json:"component_name"`
	ComponentType string  `json:"component_type,omitempty"`
	IsNew         bool    `json:"is_new,omitempty"`
	Status        string  `json:"status"`
	Skipped       bool    `json:"skipped,omitempty"`
	CacheStatus   string  `json:"cache_status,omitempty"` // deprecated: use change_reason in UI
	ChangeReason  string  `json:"change_reason,omitempty"`
	ImageDigest   string  `json:"image_digest,omitempty"` // sha256:...
	Duration      float64 `json:"duration,omitempty"`     // seconds
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.NoConfigChanges && !run.Force {
		l.Info("no config changes, skipping builds")
		return nil
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	isPreview := run.IsPreview()

	// Get app config with component IDs
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	l.Info("triggering builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	if len(appConfig.ComponentIDs) == 0 {
		l.Info("no components to build")
		s.markBuildsCompleted(ctx, l, true)
		return nil
	}

	// Determine previous run's app config for build diffing via comparison baseline.
	var previousAppConfigID string
	if run.Comparison != nil && run.Comparison.BaseRun != nil && run.Comparison.BaseRun.AppConfigID != "" {
		previousAppConfigID = run.Comparison.BaseRun.AppConfigID
	} else if run.Comparison != nil && run.Comparison.BaseRunID != nil && *run.Comparison.BaseRunID != "" {
		prevRun, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, *run.Comparison.BaseRunID)
		if err == nil && prevRun.AppConfigID != "" {
			previousAppConfigID = prevRun.AppConfigID
		}
	}

	builds, err := s.buildComponents(ctx, l, appConfig, run.AppConfigID, previousAppConfigID, run.Force)
	if err != nil {
		s.markBuildsCompleted(ctx, l, false)
		s.finalizeBuildMetadata(ctx, builds, false)
		if isPreview && run.PRNumber != nil {
			s.finalizePreview(ctx, l, run, err)
		}
		return fmt.Errorf("component builds failed: %w", err)
	}

	ociArtifacts, err := activities.AwaitOrgHasFeature(ctx, activities.OrgHasFeatureRequest{
		OrgID:   run.OrgID,
		Feature: string(app.OrgFeatureSandboxOCIArtifacts),
	})
	if err != nil {
		return fmt.Errorf("unable to check sandbox-oci-artifacts feature flag: %w", err)
	}

	if ociArtifacts {
		sandboxEntry := buildEntry{
			ComponentID:   "sandbox",
			ComponentName: "Sandbox",
			ComponentType: "sandbox",
			Status:        "in-progress",
			ChangeReason:  activities.ChangeReasonSourceChanged,
		}
		builds = append(builds, sandboxEntry)
		s.updateBuildMetadata(ctx, builds)

		if err := s.buildSandbox(ctx, l); err != nil {
			s.setBuildStatus(builds, "sandbox", "error")
			s.markBuildsCompleted(ctx, l, false)
			s.finalizeBuildMetadata(ctx, builds, false)
			if isPreview && run.PRNumber != nil {
				s.finalizePreview(ctx, l, run, err)
			}
			return fmt.Errorf("sandbox build failed: %w", err)
		}
		s.setBuildStatus(builds, "sandbox", "success")
		s.updateBuildMetadata(ctx, builds)
	} else {
		l.Info("sandbox-oci-artifacts disabled, skipping sandbox build")
	}

	if isPreview && run.PRNumber != nil {
		s.finalizePreview(ctx, l, run, nil)
	}

	s.markBuildsCompleted(ctx, l, true)
	s.finalizeBuildMetadata(ctx, builds, true)
	l.Info("all builds completed successfully")
	return nil
}

// buildComponents enqueues queuebuild signals directly to component queues
// (batched, parallel within each batch) and tracks progress via the parent
// step's status metadata — no sub-steps or sub-groups are created.
func (s *Signal) buildComponents(ctx workflow.Context, l log.Logger, appConfig *app.AppConfig, appConfigID, previousAppConfigID string, force bool) ([]buildEntry, error) {
	componentIDs := appConfig.ComponentIDs

	type componentInfo struct {
		Name string
		Type string
	}
	components := make(map[string]componentInfo, len(componentIDs))
	for _, componentID := range componentIDs {
		cmp, err := activities.AwaitGetComponentByIDByComponentID(ctx, componentID)
		if err == nil {
			components[componentID] = componentInfo{Name: cmp.Name, Type: string(cmp.Type)}
		}
	}

	sourceChangedByName := map[string]bool{}
	if previousAppConfigID != "" {
		sourceOut, err := activities.AwaitLoadComparisonSourceChanged(ctx, &activities.LoadComparisonSourceChangedInput{
			RunID: s.RunID,
		})
		if err == nil && sourceOut != nil && sourceOut.ByComponentName != nil {
			sourceChangedByName = sourceOut.ByComponentName
		}
	}

	builds := make([]buildEntry, 0, len(componentIDs))
	var toBuild []string
	for _, componentID := range componentIDs {
		info := components[componentID]
		name := info.Name
		if name == "" {
			name = componentID
		}

		isNew := previousAppConfigID == ""

		if previousAppConfigID != "" && !force {
			checkInput := &activities.CheckBuildNeededInput{
				ComponentID:    componentID,
				NewAppConfigID: appConfigID,
				OldAppConfigID: previousAppConfigID,
			}
			if _, ok := sourceChangedByName[name]; ok {
				changed := sourceChangedByName[name]
				checkInput.SourceChanged = &changed
			}

			check, err := activities.AwaitCheckBuildNeeded(ctx, checkInput)
			if err == nil && !check.NeedsBuild {
				l.Info("skipping build for unchanged component",
					"component_id", componentID,
					"existing_build_id", check.ExistingBuildID)
				builds = append(builds, buildEntry{
					BuildID:       check.ExistingBuildID,
					ComponentID:   componentID,
					ComponentName: name,
					ComponentType: info.Type,
					IsNew:         false,
					Status:        "skipped",
					Skipped:       true,
					ChangeReason:  activities.ChangeReasonNoChanges,
				})
				continue
			}
			if err == nil && check.NeedsBuild {
				builds = append(builds, buildEntry{
					ComponentID:   componentID,
					ComponentName: name,
					ComponentType: info.Type,
					IsNew:         isNew,
					Status:        "pending",
					ChangeReason:  check.ChangeReason,
				})
				toBuild = append(toBuild, componentID)
				continue
			}
			if err != nil {
				isNew = true
			}
		}

		builds = append(builds, buildEntry{
			ComponentID:   componentID,
			ComponentName: name,
			ComponentType: info.Type,
			IsNew:         isNew,
			Status:        "pending",
			ChangeReason:  activities.ChangeReasonSourceChanged,
		})
		toBuild = append(toBuild, componentID)
	}

	// Update parent step metadata with initial build list
	s.updateBuildMetadata(ctx, builds)

	if len(toBuild) == 0 {
		l.Info("all components unchanged, no builds needed")
		return builds, nil
	}

	// Phase 1: Enqueue all queuebuild signals in batches. Each queuebuild
	// creates a build record and enqueues the actual build signal onto the
	// component queue. We await queuebuild completion (via callback) so we
	// know the build records exist before moving to phase 2.
	for batchStart := 0; batchStart < len(toBuild); batchStart += buildBatchSize {
		batchEnd := batchStart + buildBatchSize
		if batchEnd > len(toBuild) {
			batchEnd = len(toBuild)
		}
		batch := toBuild[batchStart:batchEnd]

		l.Info("dispatching build batch",
			"batch_start", batchStart+1,
			"batch_end", batchEnd,
			"count", len(batch))

		for _, componentID := range batch {
			s.setBuildStatus(builds, componentID, "in-progress")
		}
		s.updateBuildMetadata(ctx, builds)

		type pendingEnqueue struct {
			componentID string
			cb          callback.Ref
		}
		pending := make([]pendingEnqueue, 0, len(batch))

		for _, componentID := range batch {
			cb := callback.New(ctx, fmt.Sprintf("enqueue-%s", componentID))
			_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
				OwnerID:         componentID,
				OwnerType:       "components",
				SignalOwnerID:   componentID,
				SignalOwnerType: "components",
				Signal: &queuebuild.Signal{
					ComponentID:    componentID,
					AppConfigID:    appConfigID,
					AppBranchRunID: s.RunID,
				},
				Callback: cb,
			})
			if err != nil {
				s.setBuildStatus(builds, componentID, "error")
				s.updateBuildMetadata(ctx, builds)
				return builds, fmt.Errorf("component %s: enqueue failed: %w", componentID, err)
			}
			pending = append(pending, pendingEnqueue{componentID: componentID, cb: cb})
		}

		for _, p := range pending {
			if _, err := callback.AwaitWithTimeout(ctx, p.cb, callback.FallbackAwaitTimeout); err != nil {
				s.setBuildStatus(builds, p.componentID, "error")
				s.updateBuildMetadata(ctx, builds)
				return builds, fmt.Errorf("component %s: queuebuild failed: %w", p.componentID, err)
			}
		}
	}

	// Phase 2: All build records exist and build signals are enqueued. Poll
	// until every build reaches a terminal status.
	l.Info("all builds enqueued, awaiting build completions", "count", len(toBuild))

	for {
		result, err := activities.AwaitCheckBuildsCompleteByRunID(ctx, s.RunID)
		if err != nil {
			return builds, fmt.Errorf("unable to check build statuses: %w", err)
		}

		if result.AllDone {
			for _, br := range result.Builds {
				s.setBuildID(builds, br.ComponentID, br.BuildID)
				if br.Status == string(app.ComponentBuildStatusActive) {
					s.setBuildStatus(builds, br.ComponentID, "success")
				} else {
					s.setBuildStatus(builds, br.ComponentID, "error")
				}
			}
			s.updateBuildMetadata(ctx, builds)

			if result.HasError {
				return builds, fmt.Errorf("one or more component builds failed")
			}
			break
		}

		// Update metadata with current statuses
		for _, br := range result.Builds {
			s.setBuildID(builds, br.ComponentID, br.BuildID)
			switch br.Status {
			case string(app.ComponentBuildStatusActive):
				s.setBuildStatus(builds, br.ComponentID, "success")
			case string(app.ComponentBuildStatusError), "cancelled":
				s.setBuildStatus(builds, br.ComponentID, "error")
			}
		}
		s.updateBuildMetadata(ctx, builds)

		// Sleep before polling again
		if err := workflow.Sleep(ctx, 5*time.Second); err != nil {
			return builds, fmt.Errorf("sleep interrupted: %w", err)
		}
	}

	l.Info("all component builds completed successfully")
	return builds, nil
}

// setBuildStatus updates a build entry's status in the builds list.
func (s *Signal) setBuildStatus(builds []buildEntry, componentID, status string) {
	for i := range builds {
		if builds[i].ComponentID == componentID {
			builds[i].Status = status
			return
		}
	}
}

func (s *Signal) setBuildID(builds []buildEntry, componentID, buildID string) {
	for i := range builds {
		if builds[i].ComponentID == componentID {
			builds[i].BuildID = buildID
			return
		}
	}
}

// updateBuildMetadata writes the current builds list to the parent step's
// status metadata so the UI can display real-time build progress.
func (s *Signal) updateBuildMetadata(ctx workflow.Context, builds []buildEntry) {
	s.updateBuildMetadataWithCompleted(ctx, builds, nil)
}

func (s *Signal) updateBuildMetadataWithCompleted(ctx workflow.Context, builds []buildEntry, buildsCompleted *bool) {
	if s.StepID == "" {
		return
	}

	buildList := make([]any, 0, len(builds))
	for _, b := range builds {
		entry := map[string]any{
			"component_id":   b.ComponentID,
			"component_name": b.ComponentName,
			"component_type": b.ComponentType,
			"is_new":         b.IsNew,
			"status":         b.Status,
			"skipped":        b.Skipped,
			"cache_status":   b.CacheStatus,
		}
		if b.ChangeReason != "" {
			entry["change_reason"] = b.ChangeReason
		}
		if b.BuildID != "" {
			entry["build_id"] = b.BuildID
		}
		if b.ImageDigest != "" {
			entry["image_digest"] = b.ImageDigest
		}
		if b.Duration > 0 {
			entry["duration"] = b.Duration
		}
		buildList = append(buildList, entry)
	}

	meta := map[string]any{
		"builds": buildList,
	}
	statusVal := app.StatusInProgress
	desc := fmt.Sprintf("building %d components", len(builds))
	if buildsCompleted != nil {
		meta[app.AppBranchRunLabelBuildsCompleted] = *buildsCompleted
		if *buildsCompleted {
			statusVal = app.StatusSuccess
			desc = fmt.Sprintf("built %d components", len(builds))
		} else {
			statusVal = app.StatusError
			desc = "builds failed"
		}
	}

	status := app.CompositeStatus{
		Status:                 statusVal,
		StatusHumanDescription: desc,
		Metadata:               meta,
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     s.StepID,
		Status: status,
	})
}

func (s *Signal) markBuildsCompleted(ctx workflow.Context, l log.Logger, completed bool) {
	if err := activities.AwaitUpdateAppBranchRunBuildsCompleted(ctx, &activities.UpdateAppBranchRunBuildsCompletedInput{
		RunID:           s.RunID,
		BuildsCompleted: completed,
	}); err != nil {
		l.Warn("unable to update builds_completed label", "error", err, "builds_completed", completed)
	}
}

func (s *Signal) finalizeBuildMetadata(ctx workflow.Context, builds []buildEntry, completed bool) {
	s.updateBuildMetadataWithCompleted(ctx, builds, &completed)
}

func (s *Signal) finalizePreview(ctx workflow.Context, l log.Logger, run *app.AppBranchRun, buildErr error) {
	if run.PRNumber == nil {
		return
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		l.Warn("unable to get branch for preview finalization", "error", err)
		return
	}

	var vcsConfigID string
	if len(branch.Configs) > 0 {
		if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		}
	}
	if vcsConfigID == "" {
		return
	}

	status := activities.PRCommentStatusSuccess
	commitState := "success"
	commitDesc := "Preview complete"
	var errMsg string

	if buildErr != nil {
		status = activities.PRCommentStatusFailed
		commitState = "failure"
		commitDesc = "Preview failed"
		errMsg = buildErr.Error()
	}

	var diff *activities.ComputeAppConfigDiffOutput
	if run.AppConfigID != "" {
		baseline, baselineErr := activities.AwaitFindLatestNonPreviewAppConfig(ctx, &activities.FindLatestNonPreviewAppConfigInput{
			AppID: branch.AppID,
		})
		var oldConfigID string
		if baselineErr == nil && baseline.AppConfigID != "" {
			oldConfigID = baseline.AppConfigID
		}

		computed, diffErr := activities.AwaitComputeAppConfigDiff(ctx, &activities.ComputeAppConfigDiffInput{
			AppID:       branch.AppID,
			NewConfigID: run.AppConfigID,
			OldConfigID: oldConfigID,
		})
		if diffErr == nil {
			diff = computed
		}
	}

	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		AppName:      branch.Name,
		RunID:        s.RunID,
		Status:       status,
		Diff:         diff,
		ErrorMessage: errMsg,
	})

	_, _ = activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		ExistingCommentID: run.GithubCommentID,
		Body:              commentBody,
	})

	if run.HeadSHA != "" {
		_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			State:       commitState,
			Context:     "nuon/preview",
			Description: commitDesc,
		})
	}
}

func (s *Signal) buildSandbox(ctx workflow.Context, l log.Logger) error {
	cb := callback.New(ctx, s.AppBranchID+"-sandbox-infra")
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.AppBranchID,
		OwnerType:       "app_branches",
		QueueName:       "app-branch-sandbox-builds",
		SignalOwnerID:   s.AppBranchID,
		SignalOwnerType: "app_branches",
		Signal: &sandboxbuild.Signal{
			AppBranchID: s.AppBranchID,
			RunID:       s.RunID,
			StepID:      s.StepID,
		},
		Callback: cb,
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue sandbox build signal: %w", err)
	}

	if _, err = callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout); err != nil {
		return fmt.Errorf("sandbox build failed: %w", err)
	}

	l.Info("sandbox infrastructure build completed")
	return nil
}
