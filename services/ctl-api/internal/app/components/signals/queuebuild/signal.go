package queuebuild

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Signal struct {
	ComponentID    string `json:"component_id" validate:"required"`
	AppConfigID    string `json:"app_config_id"`
	BuildID        string `json:"build_id"`
	AppBranchRunID string `json:"app_branch_run_id"`
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithAutoRetry      = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	buildID := s.BuildID

	if buildID == "" && s.AppConfigID != "" {
		adopted, err := activities.AwaitAdoptQueuedComponentBuild(ctx, activities.AdoptQueuedComponentBuildRequest{
			ComponentID:    s.ComponentID,
			AppConfigID:    s.AppConfigID,
			AppBranchRunID: s.AppBranchRunID,
		})
		if err != nil {
			return fmt.Errorf("unable to adopt queued build: %w", err)
		}
		buildID = adopted.BuildID
	}

	if buildID == "" {
		cmp, err := activities.AwaitGetComponentByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component: %w", err)
		}

		req := activities.CreateComponentBuildRecordRequest{
			ComponentID:    s.ComponentID,
			OrgID:          cmp.OrgID,
			AppBranchRunID: s.AppBranchRunID,
			AppConfigID:    s.AppConfigID,
		}

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

	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.ComponentID,
		OwnerType: "components",
		QueueName: queuenames.ComponentDefaultQueueName,
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
