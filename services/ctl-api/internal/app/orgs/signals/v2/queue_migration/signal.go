package queue_migration

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-queue-migration"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Mark migration started
	if err := activities.AwaitUpdateOrgStatusV2Metadata(ctx, activities.UpdateOrgStatusV2MetadataRequest{
		OrgID: s.OrgID,
		Data: map[string]any{
			"queue_migration_started_at":  time.Now().Unix(),
			"queue_migration_finished_at": nil,
		},
	}); err != nil {
		l.Error("unable to update migration started metadata", zap.Error(err))
	}

	// 1. Ensure org-signals queue
	l.Info("ensuring org-signals queue", zap.String("org_id", s.OrgID))
	if err := activities.AwaitEnsureOrgQueueByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("unable to ensure org queue: %w", err)
	}

	// 2. Ensure app queues (branches + components)
	apps, err := activities.AwaitGetOrgAppsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org apps: %w", err)
	}

	for _, app := range apps {
		l.Info("ensuring app queues", zap.String("app_id", app.ID))
		if err := activities.AwaitEnsureAppQueueByAppID(ctx, app.ID); err != nil {
			l.Warn("unable to ensure app queues", zap.String("app_id", app.ID), zap.Error(err))
		}

		l.Info("ensuring app branch queues", zap.String("app_id", app.ID))

		branches, err := activities.AwaitGetAppBranchesByAppID(ctx, app.ID)
		if err != nil {
			return fmt.Errorf("unable to get app branches for %s: %w", app.ID, err)
		}
		for _, branch := range branches {
			if err := activities.AwaitEnsureAppBranchQueueByBranchID(ctx, branch.ID); err != nil {
				l.Warn("unable to ensure app branch queue", zap.String("branch_id", branch.ID), zap.Error(err))
			}
		}

		l.Info("ensuring component queues", zap.String("app_id", app.ID))

		components, err := activities.AwaitGetAppComponentsByAppID(ctx, app.ID)
		if err != nil {
			return fmt.Errorf("unable to get app components for %s: %w", app.ID, err)
		}
		for _, component := range components {
			if err := activities.AwaitEnsureComponentQueueByComponentID(ctx, component.ID); err != nil {
				l.Warn("unable to ensure component queue", zap.String("component_id", component.ID), zap.Error(err))
			}
		}
	}

	// 3. Ensure install queues
	installs, err := activities.AwaitGetOrgInstallsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org installs: %w", err)
	}

	for _, install := range installs {
		l.Info("ensuring install queues", zap.String("install_id", install.ID))
		if err := activities.AwaitEnsureInstallQueuesByInstallID(ctx, install.ID); err != nil {
			l.Warn("unable to ensure install queues", zap.String("install_id", install.ID), zap.Error(err))
		}

		// Ensure install-level runner queues
		installRunners, err := activities.AwaitGetInstallRunnersByInstallID(ctx, install.ID)
		if err != nil {
			l.Warn("unable to get install runners", zap.String("install_id", install.ID), zap.Error(err))
			continue
		}
		for _, runner := range installRunners {
			l.Info("ensuring install runner queues", zap.String("runner_id", runner.ID), zap.String("install_id", install.ID))
			if err := activities.AwaitEnsureRunnerQueuesByRunnerID(ctx, runner.ID); err != nil {
				l.Warn("unable to ensure install runner queues", zap.String("runner_id", runner.ID), zap.Error(err))
			}
		}
	}

	// 4. Ensure runner queues
	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	for _, runner := range org.RunnerGroup.Runners {
		l.Info("ensuring runner queues", zap.String("runner_id", runner.ID))
		if err := activities.AwaitEnsureRunnerQueuesByRunnerID(ctx, runner.ID); err != nil {
			l.Warn("unable to ensure runner queues", zap.String("runner_id", runner.ID), zap.Error(err))
		}
	}

	// 5. Ensure VCS connection queues
	vcsConns, err := activities.AwaitGetOrgVCSConnectionsByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get vcs connections: %w", err)
	}

	for _, conn := range vcsConns {
		l.Info("ensuring vcs connection queue", zap.String("vcs_connection_id", conn.ID))
		if err := activities.AwaitEnsureVCSConnectionQueueByVCSConnectionID(ctx, conn.ID); err != nil {
			l.Warn("unable to ensure vcs connection queue", zap.String("vcs_connection_id", conn.ID), zap.Error(err))
		}
	}

	// 6. Enable queues feature flag
	l.Info("enabling queues feature flag", zap.String("org_id", s.OrgID))
	if err := activities.AwaitEnableQueuesFeatureFlagByOrgID(ctx, s.OrgID); err != nil {
		return fmt.Errorf("unable to enable queues feature flag: %w", err)
	}

	// Mark migration finished
	if err := activities.AwaitUpdateOrgStatusV2Metadata(ctx, activities.UpdateOrgStatusV2MetadataRequest{
		OrgID: s.OrgID,
		Data: map[string]any{
			"queue_migration_finished_at": time.Now().Unix(),
		},
	}); err != nil {
		l.Error("unable to update migration finished metadata", zap.Error(err))
	}

	l.Info("queue migration complete", zap.String("org_id", s.OrgID))
	return nil
}
