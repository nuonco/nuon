package builds

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/sandboxbuild"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

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
	if appConfig.BuildsLocked {
		_, err := activities.AwaitUpdateAppBranchRunBuildsCompleted(ctx, &activities.UpdateAppBranchRunBuildsCompletedInput{
			RunID:           s.RunID,
			BuildsCompleted: true,
		})
		return err
	}

	l.Info("triggering builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

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

	builds, err := s.buildsAsReady(ctx, l, appConfig, run.AppConfigID, previousAppConfigID, run.Force)
	if err != nil {
		buildsErr := s.markBuildsCompleted(ctx, l, false)
		s.finalizeBuildMetadata(ctx, builds, false, buildsErr)
		if isPreview && run.PRNumber != nil {
			s.finalizePreview(ctx, l, run, builds, err)
		}
		return buildsFailure("builds failed", buildsErr, err)
	}
	return s.finishBuilds(ctx, l, run, appConfig, builds)
}

func (s *Signal) finishBuilds(ctx workflow.Context, l log.Logger, run *app.AppBranchRun, appConfig *app.AppConfig, builds []buildEntry) error {
	isPreview := run.IsPreview()
	if isPreview && run.PRNumber != nil {
		s.finalizePreview(ctx, l, run, builds, nil)
	}

	if appConfig.BuildsPinned {
		if _, err := activities.AwaitUpdateAppBranchRunBuildsCompleted(ctx, &activities.UpdateAppBranchRunBuildsCompletedInput{
			RunID:           s.RunID,
			BuildsCompleted: true,
		}); err != nil {
			buildsErr := s.markBuildsCompleted(ctx, l, false)
			s.finalizeBuildMetadata(ctx, builds, false, buildsErr)
			if isPreview && run.PRNumber != nil {
				s.finalizePreview(ctx, l, run, builds, err)
			}
			return fmt.Errorf("unable to finalize selected builds: %w", err)
		}
	} else {
		_ = s.markBuildsCompleted(ctx, l, true)
	}
	s.finalizeBuildMetadata(ctx, builds, true, nil)
	l.Info("all builds completed successfully")
	return nil
}

// buildsFailure returns the step error for a failed builds step, preferring the
// canonical composite error message so the step description names the actual
// cause instead of the generic wrapper.
func buildsFailure(wrapper string, ce *compositeerrors.CompositeErrorData, cause error) error {
	if ce != nil && ce.Message != "" {
		return temporal.NewNonRetryableApplicationError(ce.Message, string(ce.Type), nil)
	}
	return fmt.Errorf("%s: %w", wrapper, cause)
}

// updateBuildMetadata writes the current builds list to the parent step's
// status metadata so the UI can display real-time build progress.
func (s *Signal) updateBuildMetadata(ctx workflow.Context, builds []buildEntry) {
	s.updateBuildMetadataWithCompleted(ctx, builds, nil, nil)
}

func (s *Signal) updateBuildMetadataWithCompleted(ctx workflow.Context, builds []buildEntry, buildsCompleted *bool, ce *compositeerrors.CompositeErrorData) {
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
			if ce != nil && ce.Message != "" {
				desc = ce.Message
			}
		}
	}

	status := app.CompositeStatus{
		Status:                 statusVal,
		StatusHumanDescription: desc,
		CompositeError:         ce,
		Metadata:               meta,
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     s.StepID,
		Status: status,
	})
}

func (s *Signal) markBuildsCompleted(ctx workflow.Context, l log.Logger, completed bool) *compositeerrors.CompositeErrorData {
	out, err := activities.AwaitUpdateAppBranchRunBuildsCompleted(ctx, &activities.UpdateAppBranchRunBuildsCompletedInput{
		RunID:           s.RunID,
		BuildsCompleted: completed,
	})
	if err != nil {
		l.Warn("unable to update builds_completed label", "error", err, "builds_completed", completed)
		return nil
	}
	if completed || out == nil {
		return nil
	}
	return out.CompositeError
}

func (s *Signal) finalizeBuildMetadata(ctx workflow.Context, builds []buildEntry, completed bool, ce *compositeerrors.CompositeErrorData) {
	s.updateBuildMetadataWithCompleted(ctx, builds, &completed, ce)
}

func (s *Signal) finalizePreview(ctx workflow.Context, l log.Logger, run *app.AppBranchRun, builds []buildEntry, buildErr error) {
	if run.PRNumber == nil || !run.PreviewGitHubComment() {
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
	var errMsg string
	if buildErr != nil {
		status = activities.PRCommentStatusFailed
		errMsg = buildErr.Error()
	}

	var diff *activities.ComputeAppConfigDiffOutput
	if run.AppConfigID != "" {
		baseline, baselineErr := activities.AwaitResolvePreviewBaselineAppConfig(ctx, &activities.ResolvePreviewBaselineAppConfigInput{
			RunID:       s.RunID,
			AppBranchID: s.AppBranchID,
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

	commentContext, _ := activities.AwaitGetPreviewCommentContext(ctx, &activities.GetPreviewCommentContextInput{
		RunID: s.RunID,
	})

	// Derive phases from DB context and override Builds with the known result.
	// Config is accurately derived from the DB (the appconfig step completed
	// before builds started).  Builds is still in-progress in the DB at this
	// point, so we override it explicitly.
	phases := commentContextPhases(commentContext)
	if buildErr != nil {
		phases.Builds = activities.PRCommentPhaseInvalid
	} else {
		phases.Builds = activities.PRCommentPhaseValid
	}

	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		OrgName:          branch.Org.Name,
		AppName:          branch.App.Name,
		AppBranchID:      branch.ID,
		BranchName:       branch.Name,
		RunID:            s.RunID,
		RunURL:           previewRunURL(commentContext),
		Status:           status,
		Mode:             run.PreviewMode(),
		Diff:             diff,
		ComponentChanges: componentBuildChanges(builds, commentContext),
		// PreviewInstallName and InstallApplied are intentionally omitted:
		// the install step has not run yet at this point.  The run finalizer
		// writes the final comment once the install step completes.
		ErrorMessage: errMsg,
		Phases:       phases,
	})

	_, _ = activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		AppBranchID:       run.AppBranchID,
		ExistingCommentID: run.GithubCommentID,
		Body:              commentBody,
	})
}

func componentBuildChanges(builds []buildEntry, commentContext *activities.GetPreviewCommentContextOutput) []activities.ComponentBuildChange {
	buildURLs := make(map[string]string)
	if commentContext != nil {
		for _, change := range commentContext.ComponentChanges {
			buildURLs[change.ComponentID] = change.BuildURL
		}
	}

	changes := make([]activities.ComponentBuildChange, 0, len(builds))
	for _, build := range builds {
		if build.ChangeReason != activities.ChangeReasonSourceChanged &&
			build.ChangeReason != activities.ChangeReasonConfigChanged {
			continue
		}
		changes = append(changes, activities.ComponentBuildChange{
			ComponentName: build.ComponentName,
			ComponentID:   build.ComponentID,
			BuildID:       build.BuildID,
			ChangeReason:  build.ChangeReason,
			BuildURL:      buildURLs[build.ComponentID],
		})
	}
	return changes
}

func previewRunURL(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.RunURL
}

// commentContextPhases returns a copy of the Phases from commentContext, or an
// empty PRCommentPhases when the context is nil or carries no phase data.
func commentContextPhases(commentContext *activities.GetPreviewCommentContextOutput) *activities.PRCommentPhases {
	if commentContext == nil || commentContext.Phases == nil {
		return &activities.PRCommentPhases{}
	}
	cp := *commentContext.Phases
	return &cp
}

func previewInstallName(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.PreviewInstallName
}

func previewInstallURL(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.PreviewInstallURL
}

func (s *Signal) buildSandbox(ctx workflow.Context, l log.Logger) error {
	cb := callback.New(ctx, s.AppBranchID+"-sandbox-infra")
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.AppBranchID,
		OwnerType:       "app_branches",
		QueueName:       queuenames.AppBranchSandboxBuildsQueueName,
		SignalOwnerID:   s.AppBranchID,
		SignalOwnerType: "app_branches",
		Signal: &sandboxbuild.Signal{
			AppBranchID:  s.AppBranchID,
			RunID:        s.RunID,
			StepID:       s.StepID,
			ReuseChecked: true,
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
