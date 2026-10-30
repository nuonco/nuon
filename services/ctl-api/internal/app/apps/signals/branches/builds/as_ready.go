package builds

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/queuebuild"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const componentSelectionConcurrency = 5

func (s *Signal) buildsAsReady(ctx workflow.Context, l log.Logger, appConfig *app.AppConfig, appConfigID, previousAppConfigID string, force bool) ([]buildEntry, error) {
	builds := make([]buildEntry, len(appConfig.ComponentIDs)+1)
	for i, componentID := range appConfig.ComponentIDs {
		builds[i] = buildEntry{
			ComponentID:   componentID,
			ComponentName: componentID,
			IsNew:         previousAppConfigID == "",
			Status:        "pending",
			ChangeReason:  activities.ChangeReasonSourceChanged,
		}
	}
	builds[len(appConfig.ComponentIDs)] = buildEntry{
		ComponentID:   activities.SandboxComponentID,
		ComponentName: "Sandbox",
		ComponentType: "sandbox",
		Status:        "pending",
	}

	dirty := true
	changed := func() { dirty = true }
	remaining := 3
	var componentErr, sandboxErr, comparisonErr error
	workflow.Go(ctx, func(ctx workflow.Context) {
		componentErr = s.componentsAsReady(ctx, l, builds[:len(appConfig.ComponentIDs)], appConfigID, previousAppConfigID, force, changed)
		remaining--
		changed()
	})
	workflow.Go(ctx, func(ctx workflow.Context) {
		sandboxErr = s.sandboxAsReady(ctx, l, &builds[len(appConfig.ComponentIDs)], appConfigID, previousAppConfigID, force, changed)
		remaining--
		changed()
	})
	workflow.Go(ctx, func(ctx workflow.Context) {
		_, err := activities.AwaitComputeAndStoreAppBranchRunComparison(ctx, &activities.ComputeAndStoreAppBranchRunComparisonInput{
			AppBranchID:         s.AppBranchID,
			RunID:               s.RunID,
			ExternalSourcesOnly: true,
		})
		if err != nil {
			comparisonErr = fmt.Errorf("unable to enrich external source comparison: %w", err)
		}
		remaining--
		changed()
	})

	for remaining > 0 || dirty {
		if dirty {
			dirty = false
			s.updateBuildMetadata(ctx, builds)
		}
		if err := workflow.Await(ctx, func() bool { return dirty || remaining == 0 }); err != nil {
			return builds, err
		}
	}
	return builds, errors.Join(componentErr, sandboxErr, comparisonErr)
}

func (s *Signal) componentsAsReady(ctx workflow.Context, l log.Logger, builds []buildEntry, appConfigID, previousAppConfigID string, force bool, changed func()) error {
	if len(builds) == 0 {
		return nil
	}
	sourceChangedByName := map[string]bool{}
	if previousAppConfigID != "" && !force {
		sourceOut, err := activities.AwaitLoadComparisonSourceChanged(ctx, &activities.LoadComparisonSourceChangedInput{RunID: s.RunID})
		if err == nil && sourceOut != nil && sourceOut.ByComponentName != nil {
			sourceChangedByName = sourceOut.ByComponentName
		}
	}

	next, completed := 0, 0
	selectionErrors := make([]error, len(builds))
	workers := min(componentSelectionConcurrency, len(builds))
	for range workers {
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer func() { completed++ }()
			for next < len(builds) && ctx.Err() == nil {
				i := next
				next++
				if err := s.selectAndEnqueueComponent(ctx, l, &builds[i], appConfigID, previousAppConfigID, force, sourceChangedByName, changed); err != nil {
					builds[i].Status = "error"
					selectionErrors[i] = err
					changed()
				}
			}
		})
	}
	if err := workflow.Await(ctx, func() bool { return completed == workers }); err != nil {
		return err
	}
	if err := errors.Join(selectionErrors...); err != nil {
		return err
	}

	hasBuilds := false
	for _, build := range builds {
		if !build.Skipped {
			hasBuilds = true
			break
		}
	}
	if !hasBuilds {
		return nil
	}
	l.Info("all builds enqueued, awaiting build completions")
	for {
		result, err := activities.AwaitCheckBuildsCompleteByRunID(ctx, s.RunID)
		if err != nil {
			return fmt.Errorf("unable to check build statuses: %w", err)
		}
		for _, br := range result.Builds {
			for i := range builds {
				if builds[i].ComponentID != br.ComponentID || builds[i].Skipped {
					continue
				}
				builds[i].BuildID = br.BuildID
				switch br.Status {
				case string(app.ComponentBuildStatusActive):
					builds[i].Status = "success"
				case string(app.ComponentBuildStatusError), "cancelled":
					builds[i].Status = "error"
				}
			}
		}
		changed()
		allDone := result.AllDone
		for _, build := range builds {
			if !build.Skipped && build.Status != "success" && build.Status != "error" {
				allDone = false
			}
		}
		if allDone {
			if result.HasError {
				return fmt.Errorf("one or more component builds failed")
			}
			return nil
		}
		if err := workflow.Sleep(ctx, 5*time.Second); err != nil {
			return fmt.Errorf("sleep interrupted: %w", err)
		}
	}
}

func (s *Signal) selectAndEnqueueComponent(ctx workflow.Context, l log.Logger, build *buildEntry, appConfigID, previousAppConfigID string, force bool, sourceChangedByName map[string]bool, changed func()) error {
	componentID := build.ComponentID
	cmp, err := activities.AwaitGetComponentByIDByComponentID(ctx, componentID)
	if err == nil && cmp != nil {
		if cmp.Name != "" {
			build.ComponentName = cmp.Name
		}
		build.ComponentType = string(cmp.Type)
		changed()
	}
	var prepared *sourceidentity.Identity
	if previousAppConfigID != "" && !force {
		input := &activities.CheckBuildNeededInput{
			ComponentID:    componentID,
			NewAppConfigID: appConfigID,
			OldAppConfigID: previousAppConfigID,
			RunID:          s.RunID,
		}
		if sourceChanged, ok := sourceChangedByName[build.ComponentName]; ok {
			input.SourceChanged = &sourceChanged
		}
		check, err := activities.AwaitCheckBuildNeeded(ctx, input)
		if err != nil {
			return fmt.Errorf("component %s: prepare source: %w", componentID, err)
		}
		prepared = check.SourceIdentity
		build.ChangeReason = check.ChangeReason
		if !check.NeedsBuild {
			if check.ExistingBuildID != "" {
				if err := activities.AwaitSetComponentConfigLatestBuild(ctx, &activities.SetComponentConfigLatestBuildInput{
					AppConfigID: appConfigID,
					ComponentID: componentID,
					BuildID:     check.ExistingBuildID,
				}); err != nil {
					return fmt.Errorf("component %s: pin latest build: %w", componentID, err)
				}
			}
			build.BuildID = check.ExistingBuildID
			build.Status = "skipped"
			build.Skipped = true
			build.ChangeReason = activities.ChangeReasonNoChanges
			changed()
			l.Info("skipping build for unchanged component", "component_id", componentID, "existing_build_id", check.ExistingBuildID)
			return nil
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	build.Status = "in-progress"
	changed()
	cb := callback.New(ctx, fmt.Sprintf("enqueue-%s", componentID))
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         componentID,
		OwnerType:       "components",
		QueueName:       queuenames.ComponentDefaultQueueName,
		SignalOwnerID:   componentID,
		SignalOwnerType: "components",
		Signal: &queuebuild.Signal{
			ComponentID:    componentID,
			AppConfigID:    appConfigID,
			AppBranchRunID: s.RunID,
			SourceIdentity: prepared,
		},
		Callback: cb,
	})
	if err != nil {
		return fmt.Errorf("component %s: enqueue failed: %w", componentID, err)
	}
	if _, err := callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout); err != nil {
		return fmt.Errorf("component %s: queuebuild failed: %w", componentID, err)
	}
	return nil
}

func (s *Signal) sandboxAsReady(ctx workflow.Context, l log.Logger, build *buildEntry, appConfigID, previousAppConfigID string, force bool, changed func()) (err error) {
	defer func() {
		if err != nil {
			build.Status = "error"
			changed()
		}
	}()
	check, err := activities.AwaitCheckSandboxBuildNeeded(ctx, &activities.CheckSandboxBuildNeededInput{
		NewAppConfigID: appConfigID,
		OldAppConfigID: previousAppConfigID,
		RunID:          s.RunID,
		Force:          force,
	})
	if err != nil {
		return fmt.Errorf("unable to check whether sandbox build is needed: %w", err)
	}
	build.ChangeReason = check.ChangeReason
	if !check.NeedsBuild {
		build.BuildID = check.ExistingBuildID
		build.Status = "skipped"
		build.Skipped = true
		build.ChangeReason = activities.ChangeReasonNoChanges
		changed()
		return nil
	}
	build.Status = "in-progress"
	changed()
	if err := s.buildSandbox(ctx, l, check.SourceIdentity, true); err != nil {
		return err
	}
	build.Status = "success"
	changed()
	return nil
}
