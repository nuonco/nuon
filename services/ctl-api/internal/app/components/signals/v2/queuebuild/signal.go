package queuebuild

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Signal struct {
	ComponentID string `json:"component_id" validate:"required"`
	AppConfigID string `json:"app_config_id"` // optional; if set, use branch VCS commit when component shares same VCS config
	BuildID     string `json:"build_id"`      // optional; if set, skip build creation and trigger pre-created build
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	buildID := s.BuildID

	// If no pre-created build, create the build record.
	if buildID == "" {
		cmp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component: %w", err)
		}

		req := activities.CreateComponentBuildRecordRequest{
			CreatedByID: cmp.CreatedByID,
			ComponentID: s.ComponentID,
			OrgID:       cmp.OrgID,
		}

		// If AppConfigID is provided, try to use the branch's VCS commit when the
		// component shares the same VCS config as the triggering branch run.
		if s.AppConfigID != "" {
			run, err := activities.AwaitGetAppBranchRunByAppConfigIDByAppConfigID(ctx, s.AppConfigID)
			if err == nil && run.VCSConnectionCommit != nil {
				commitOwnerID := run.VCSConnectionCommit.OwnerID
				componentVCSConfigID := resolveComponentVCSConfigID(cmp)
				if componentVCSConfigID != "" && componentVCSConfigID == commitOwnerID {
					sha := run.VCSConnectionCommit.SHA
					commitID := run.VCSConnectionCommit.ID
					req.GitRef = &sha
					req.VCSConnectionCommitID = &commitID
				}
			}
		}

		build, err := activities.AwaitCreateComponentBuildRecord(ctx, req)
		if err != nil {
			return fmt.Errorf("unable to queue component build: %w", err)
		}
		buildID = build.ID
	}

	// Enqueue the build signal to the component's queue. The caller awaits this
	// queuebuild signal; the build signal runs independently on the component queue.
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.ComponentID,
		OwnerType: "components",
		Signal: &buildsignal.Signal{
			ComponentID: s.ComponentID,
			BuildID:     buildID,
		},
		SignalOwnerID:   buildID,
		SignalOwnerType: "component_builds",
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue build signal: %w", err)
	}

	return nil
}

// resolveComponentVCSConfigID returns the VCS config ID from the component's latest config,
// used to compare against the branch run's VCS commit owner.
func resolveComponentVCSConfigID(cmp *app.Component) string {
	if cmp.LatestConfig == nil {
		return ""
	}
	if cfg := cmp.LatestConfig.ConnectedGithubVCSConfig; cfg != nil {
		return cfg.ID
	}
	if cfg := cmp.LatestConfig.PublicGitVCSConfig; cfg != nil {
		return cfg.ID
	}
	return ""
}
