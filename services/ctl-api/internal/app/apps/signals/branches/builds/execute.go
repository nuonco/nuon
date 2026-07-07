package builds

import (
	"fmt"

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
	ComponentID   string  `json:"component_id"`
	ComponentName string  `json:"component_name"`
	ComponentType string  `json:"component_type,omitempty"`
	IsNew         bool    `json:"is_new,omitempty"`
	Status        string  `json:"status"`
	Skipped       bool    `json:"skipped,omitempty"`
	CacheStatus   string  `json:"cache_status,omitempty"` // "cache hit", "partial cache", "no cache"
	ImageDigest   string  `json:"image_digest,omitempty"` // sha256:...
	Duration      float64 `json:"duration,omitempty"`     // seconds
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Load the run to get the AppConfigID (set by the appconfig step)
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

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
		return nil
	}

	// Determine previous run's app config for build diffing
	var previousAppConfigID string
	if run.PreviousRunID != nil && *run.PreviousRunID != "" {
		prevRun, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, *run.PreviousRunID)
		if err == nil && prevRun.AppConfigID != "" {
			previousAppConfigID = prevRun.AppConfigID
		}
	}

	// Build all components — tracks progress via parent step metadata
	builds, err := s.buildComponents(ctx, l, appConfig, run.AppConfigID, previousAppConfigID)
	if err != nil {
		return fmt.Errorf("component builds failed: %w", err)
	}

	// Build sandbox infrastructure (terraform)
	sandboxEntry := buildEntry{
		ComponentID:   "sandbox",
		ComponentName: "Sandbox",
		ComponentType: "sandbox",
		Status:        "in-progress",
		CacheStatus:   "no cache",
	}
	builds = append(builds, sandboxEntry)
	s.updateBuildMetadata(ctx, builds)

	if err := s.buildSandbox(ctx, l); err != nil {
		s.setBuildStatus(builds, "sandbox", "error")
		s.updateBuildMetadata(ctx, builds)
		return fmt.Errorf("sandbox build failed: %w", err)
	}
	s.setBuildStatus(builds, "sandbox", "success")
	s.updateBuildMetadata(ctx, builds)

	l.Info("all builds completed successfully")
	return nil
}

// buildComponents enqueues queuebuild signals directly to component queues
// (batched, parallel within each batch) and tracks progress via the parent
// step's status metadata — no sub-steps or sub-groups are created.
func (s *Signal) buildComponents(ctx workflow.Context, l log.Logger, appConfig *app.AppConfig, appConfigID, previousAppConfigID string) ([]buildEntry, error) {
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

	builds := make([]buildEntry, 0, len(componentIDs))
	var toBuild []string
	for _, componentID := range componentIDs {
		info := components[componentID]
		name := info.Name
		if name == "" {
			name = componentID
		}

		isNew := previousAppConfigID == ""

		if previousAppConfigID != "" {
			check, err := activities.AwaitCheckBuildNeeded(ctx, &activities.CheckBuildNeededInput{
				ComponentID:    componentID,
				NewAppConfigID: appConfigID,
				OldAppConfigID: previousAppConfigID,
			})
			if err == nil && !check.NeedsBuild {
				l.Info("skipping build for unchanged component",
					"component_id", componentID,
					"existing_build_id", check.ExistingBuildID)
				builds = append(builds, buildEntry{
					ComponentID:   componentID,
					ComponentName: name,
					ComponentType: info.Type,
					IsNew:         false,
					Status:        "skipped",
					Skipped:       true,
					CacheStatus:   "cache hit",
				})
				continue
			}
			if err != nil {
				isNew = true
			}
		}

		cacheStatus := "no cache"
		if previousAppConfigID != "" {
			cacheStatus = "partial cache"
		}
		builds = append(builds, buildEntry{
			ComponentID:   componentID,
			ComponentName: name,
			ComponentType: info.Type,
			IsNew:         isNew,
			Status:        "pending",
			CacheStatus:   cacheStatus,
		})
		toBuild = append(toBuild, componentID)
	}

	// Update parent step metadata with initial build list
	s.updateBuildMetadata(ctx, builds)

	if len(toBuild) == 0 {
		l.Info("all components unchanged, no builds needed")
		return builds, nil
	}

	// Enqueue builds in batches and await each batch
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

		// Mark batch as in-progress
		for _, componentID := range batch {
			s.setBuildStatus(builds, componentID, "in-progress")
		}
		s.updateBuildMetadata(ctx, builds)

		// Enqueue all builds in this batch with callbacks
		type pendingBuild struct {
			componentID string
			cb          callback.Ref
		}
		pending := make([]pendingBuild, 0, len(batch))

		for _, componentID := range batch {
			cb := callback.New(ctx, componentID)
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
			pending = append(pending, pendingBuild{componentID: componentID, cb: cb})
		}

		// Await all builds in this batch
		var errs []error
		for _, p := range pending {
			if _, err := callback.AwaitWithTimeout(ctx, p.cb, callback.FallbackAwaitTimeout); err != nil {
				s.setBuildStatus(builds, p.componentID, "error")
				errs = append(errs, fmt.Errorf("component %s: %w", p.componentID, err))
			} else {
				s.setBuildStatus(builds, p.componentID, "success")
			}
		}
		s.updateBuildMetadata(ctx, builds)

		if len(errs) > 0 {
			return builds, fmt.Errorf("batch had %d error(s): %v", len(errs), errs)
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

// updateBuildMetadata writes the current builds list to the parent step's
// status metadata so the UI can display real-time build progress.
func (s *Signal) updateBuildMetadata(ctx workflow.Context, builds []buildEntry) {
	if s.StepID == "" {
		return
	}

	// Convert builds to a slice of map[string]any for JSON serialization
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
		if b.ImageDigest != "" {
			entry["image_digest"] = b.ImageDigest
		}
		if b.Duration > 0 {
			entry["duration"] = b.Duration
		}
		buildList = append(buildList, entry)
	}

	status := app.CompositeStatus{
		Status:                 app.StatusInProgress,
		StatusHumanDescription: fmt.Sprintf("building %d components", len(builds)),
		Metadata: map[string]any{
			"builds": buildList,
		},
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID:     s.StepID,
		Status: status,
	})
}

func (s *Signal) buildSandbox(ctx workflow.Context, l log.Logger) error {
	cb := callback.New(ctx, s.AppBranchID+"-sandbox-infra")
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.AppBranchID,
		OwnerType:       "app_branches",
		SignalOwnerID:   s.AppBranchID,
		SignalOwnerType: "app_branches",
		Signal: &sandboxbuild.Signal{
			AppBranchID: s.AppBranchID,
			RunID:       s.RunID,
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
