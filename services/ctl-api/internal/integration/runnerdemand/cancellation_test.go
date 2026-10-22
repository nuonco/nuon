package runnerdemand_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func testOwnerCancellation(t *testing.T, s testServices, namespace string, legacy *installactivities.Activities, v2 *statusactivities.Activities) {
	f := newFixture(t, s, namespace, 0, time.Minute)
	var install app.Install
	require.NoError(t, s.DB.Where(app.Install{ID: f.workflow.OwnerID}).First(&install).Error)
	var sandbox app.InstallSandbox
	require.NoError(t, s.DB.Where(app.InstallSandbox{InstallID: install.ID}).First(&sandbox).Error)
	run := &app.InstallSandboxRun{InstallID: install.ID, AppSandboxConfigID: install.AppSandboxConfigID, InstallSandboxID: &sandbox.ID, Status: app.SandboxRunStatusQueued}
	f.create(run)
	action := &app.InstallActionWorkflowRun{InstallID: install.ID, Status: app.InstallActionRunStatusQueued}
	f.create(action)

	type persistedStatus struct {
		Status    string
		StatusV2  app.CompositeStatus
		UpdatedAt time.Time
	}
	read := func(t *testing.T, owner any) persistedStatus {
		t.Helper()
		var current persistedStatus
		require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(owner).Select("status", "status_v2", "updated_at").Where(owner, "id").Take(&current).Error)
		return current
	}
	for name, domain := range map[string]struct {
		owner      any
		projection any
		finished   app.Status
		legacy     func(app.Status) error
		v2         func(app.Status) error
	}{
		"deploy": {
			owner: &app.InstallDeploy{ID: f.deploy.ID}, projection: &app.InstallComponent{ID: f.deploy.InstallComponentID}, finished: app.Status(app.InstallDeployStatusActive),
			legacy: func(status app.Status) error {
				return legacy.UpdateDeployStatus(f.ctx, installactivities.UpdateDeployStatusRequest{DeployID: f.deploy.ID, Status: app.InstallDeployStatus(status), StatusDescription: string(status)})
			},
			v2: func(status app.Status) error {
				return v2.UpdateDeployStatusV2(f.ctx, statusactivities.UpdateDeployStatusV2Request{DeployID: f.deploy.ID, Status: status, StatusDescription: string(status)})
			},
		},
		"sandbox": {
			owner: &app.InstallSandboxRun{ID: run.ID}, projection: &app.InstallSandbox{ID: sandbox.ID}, finished: app.Status(app.SandboxRunStatusActive),
			legacy: func(status app.Status) error {
				return legacy.UpdateRunStatus(f.ctx, installactivities.UpdateRunStatusRequest{RunID: run.ID, Status: app.SandboxRunStatus(status), StatusDescription: string(status)})
			},
			v2: func(status app.Status) error {
				return v2.UpdateRunStatusV2(f.ctx, statusactivities.UpdateRunStatusV2Request{RunID: run.ID, Status: app.SandboxRunStatus(status), StatusDescription: string(status)})
			},
		},
		"action": {
			owner: &app.InstallActionWorkflowRun{ID: action.ID}, finished: app.Status(app.InstallActionRunStatusFinished),
			legacy: func(status app.Status) error {
				return legacy.UpdateInstallWorkflowRunStatus(f.ctx, installactivities.UpdateInstallWorkflowRunStatusRequest{RunID: action.ID, Status: app.InstallActionWorkflowRunStatus(status), StatusDescription: string(status)})
			},
			v2: func(status app.Status) error {
				return v2.UpdateInstallWorkflowRunStatusV2(f.ctx, statusactivities.UpdateInstallWorkflowRunStatusV2Request{RunID: action.ID, Status: app.InstallActionWorkflowRunStatus(status), StatusDescription: string(status)})
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			for scenario, tc := range map[string]struct {
				completeLegacy bool
				completeV2     bool
				v2First        bool
			}{
				"completed_owner":           {completeLegacy: true, completeV2: true},
				"legacy_completion_only":    {completeLegacy: true},
				"composite_completion_only": {completeV2: true},
				"legacy_cancellation":       {},
				"composite_cancellation":    {v2First: true},
			} {
				t.Run(scenario, func(t *testing.T) {
					initial := app.NewCompositeStatus(f.ctx, app.StatusWaitingForRunner)
					initial.Metadata["origin"] = "offline-wait"
					require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(domain.owner).Updates(map[string]any{"status": app.StatusWaitingForRunner, "status_v2": initial}).Error)
					if tc.completeLegacy {
						require.NoError(t, domain.legacy(domain.finished))
					}
					if tc.completeV2 {
						require.NoError(t, domain.v2(domain.finished))
					}
					before := read(t, domain.owner)
					var projection persistedStatus
					if domain.projection != nil {
						projection = read(t, domain.projection)
					}
					first, second := domain.legacy, domain.v2
					if tc.v2First {
						first, second = second, first
					}
					require.NoError(t, first(app.StatusCancelled))
					cancelled := read(t, domain.owner)
					require.NoError(t, second(app.StatusCancelled))
					require.Equal(t, cancelled, read(t, domain.owner), "repeated cancellation must not rewrite status history")
					if tc.completeLegacy || tc.completeV2 {
						require.Equal(t, before, cancelled, "cancellation must preserve a completed domain owner")
					} else {
						require.Equal(t, string(app.StatusCancelled), cancelled.Status)
						require.Equal(t, app.StatusCancelled, cancelled.StatusV2.Status)
						require.Equal(t, "offline-wait", cancelled.StatusV2.Metadata["origin"])
						require.Len(t, cancelled.StatusV2.History, 1)
						if domain.projection != nil {
							projection = read(t, domain.projection)
							require.Equal(t, string(app.StatusCancelled), projection.Status)
							require.Equal(t, app.StatusCancelled, projection.StatusV2.Status)
						}
						require.NoError(t, domain.legacy(domain.finished))
						require.NoError(t, domain.v2(domain.finished))
						require.Equal(t, cancelled, read(t, domain.owner), "late completion must not resurrect cancelled work")
					}
					if domain.projection != nil {
						require.Equal(t, projection, read(t, domain.projection), "no-op transitions must not change the parent projection")
					}
				})
			}
		})
	}
}
