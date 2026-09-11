package activities_test

import (
	"context"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"
)

func TestWorkflowFirstStart(t *testing.T) {
	// CI discovers PostgreSQL suites using the INTEGRATION marker.
	tests.SkipIfNotIntegration(t)
	var deps struct {
		fx.In
		DB     *gorm.DB `name:"psql"`
		Seeder *testseed.Seeder
	}
	fxApp := fxtest.New(t, append(tests.CtlApiFXOptions(t), fx.Populate(&deps))...)
	fxApp.RequireStart()
	t.Cleanup(fxApp.RequireStop)
	ctx := context.Background()
	testApp := deps.Seeder.CreateApp(ctx, t)
	ctx = cctx.SetAccountContext(ctx, &testApp.CreatedBy)
	ctx = cctx.SetOrgIDContext(ctx, testApp.OrgID)
	deps.Seeder.CreateAppConfig(ctx, t, testApp.ID)
	install := deps.Seeder.CreateInstall(ctx, t, testApp)
	a := activities.New(activities.Params{DB: deps.DB})
	load := func(id string) app.Workflow {
		var wf app.Workflow
		require.NoError(t, deps.DB.WithContext(ctx).Where(app.Workflow{ID: id}).Take(&wf).Error)
		return wf
	}
	newWorkflow := func() *app.Workflow {
		return deps.Seeder.CreateWorkflow(ctx, t, install.ID, app.WorkflowTypeManualDeploy)
	}

	t.Run("concurrent starts and retries preserve first timestamp", func(t *testing.T) {
		wf := newWorkflow()
		before := load(wf.ID)
		startedAfter := time.Now().Add(-time.Millisecond)
		errs := make(chan error, 8)
		for range 8 {
			go func() {
				errs <- a.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID})
			}()
		}
		for range 8 {
			require.NoError(t, <-errs)
		}
		started := load(wf.ID)
		require.False(t, started.StartedAt.IsZero())
		require.True(t, started.StartedAt.After(startedAfter))
		require.True(t, started.StartedAt.Before(time.Now()))
		require.Equal(t, before.Status, started.Status)
		require.Equal(t, before.FinishedAt, started.FinishedAt)
		afterRestart := activities.New(activities.Params{DB: deps.DB})
		require.NoError(t, afterRestart.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID}))
		require.Equal(t, started.StartedAt, load(wf.ID).StartedAt)
	})

	t.Run("existing timestamp is not replaced", func(t *testing.T) {
		wf := newWorkflow()
		started := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
		require.NoError(t, deps.DB.WithContext(ctx).Model(wf).Update("started_at", started).Error)
		require.NoError(t, a.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID}))
		require.True(t, started.Equal(load(wf.ID).StartedAt))
	})

	t.Run("invalid and deleted IDs do not start other workflows", func(t *testing.T) {
		untouched := newWorkflow()
		deleted := newWorkflow()
		require.NoError(t, deps.DB.WithContext(ctx).Delete(deleted).Error)
		for _, id := range []string{"", "wfl00000000000000000000000", deleted.ID} {
			require.Error(t, a.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: id}))
		}
		require.True(t, load(untouched.ID).StartedAt.IsZero())
	})

	t.Run("database errors leave timestamp unset", func(t *testing.T) {
		wf := newWorkflow()
		closedTx := deps.DB.Begin()
		require.NoError(t, closedTx.Error)
		require.NoError(t, closedTx.Rollback().Error)
		failed := activities.New(activities.Params{DB: closedTx})
		require.Error(t, failed.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID}))
		require.True(t, load(wf.ID).StartedAt.IsZero())
	})
}
