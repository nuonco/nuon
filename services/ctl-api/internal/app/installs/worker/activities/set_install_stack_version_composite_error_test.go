package activities

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/stackerrors"
)

func TestSetInstallStackVersionCompositeError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_stack_versions (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
		);
		INSERT INTO install_stack_versions (id) VALUES ('sv-1');
	`).Error)

	a := &Activities{db: db}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(a.SetInstallStackVersionCompositeError)

	_, err = env.ExecuteActivity(a.SetInstallStackVersionCompositeError, SetInstallStackVersionCompositeErrorRequest{
		StackVersionID: "sv-1",
		Platform:       "aws",
		Detail:         "nested_stack_url could not be fetched: 404",
	})
	require.NoError(t, err)

	var sv app.InstallStackVersion
	require.NoError(t, db.Where(app.InstallStackVersion{ID: "sv-1"}).First(&sv).Error)
	require.NotNil(t, sv.CompositeError)
	require.Equal(t, stackerrors.StackTemplateRenderErrorType, sv.CompositeError.Type)
	require.Equal(t, "install_stack_versions", sv.CompositeError.SourceType)
	require.Equal(t, "sv-1", sv.CompositeError.SourceID)
	require.True(t, sv.CompositeError.Hints.Terminal(), "composite error must be terminal")

	_, err = env.ExecuteActivity(a.SetInstallStackVersionCompositeError, SetInstallStackVersionCompositeErrorRequest{
		StackVersionID: "sv-1",
	})
	require.NoError(t, err)

	sv = app.InstallStackVersion{}
	require.NoError(t, db.Where(app.InstallStackVersion{ID: "sv-1"}).First(&sv).Error)
	require.Nil(t, sv.CompositeError)

	_, err = env.ExecuteActivity(a.SetInstallStackVersionCompositeError, SetInstallStackVersionCompositeErrorRequest{
		StackVersionID: "missing-sv",
		Detail:         "should fail",
	})
	require.ErrorContains(t, err, gorm.ErrRecordNotFound.Error())
}

func TestSetInstallSandboxRunPlanCompositeError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE install_sandbox_runs (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			composite_error TEXT
		);
		INSERT INTO install_sandbox_runs (id) VALUES ('run-1');
	`).Error)

	a := &Activities{db: db}
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()
	env.RegisterActivity(a.SetInstallSandboxRunPlanCompositeError)

	_, err = env.ExecuteActivity(a.SetInstallSandboxRunPlanCompositeError, SetInstallSandboxRunPlanCompositeErrorRequest{
		SandboxRunID: "run-1",
		Detail:       "invalid module source path",
	})
	require.NoError(t, err)

	var run app.InstallSandboxRun
	require.NoError(t, db.Where(app.InstallSandboxRun{ID: "run-1"}).First(&run).Error)
	require.NotNil(t, run.CompositeError)
	require.Equal(t, stackerrors.SandboxPlanRenderErrorType, run.CompositeError.Type)
	require.True(t, run.CompositeError.Hints.Terminal(), "composite error must be terminal")
	require.Equal(t, "install_sandbox_runs", run.CompositeError.SourceType)

	_, err = env.ExecuteActivity(a.SetInstallSandboxRunPlanCompositeError, SetInstallSandboxRunPlanCompositeErrorRequest{
		SandboxRunID: "run-1",
	})
	require.NoError(t, err)

	run = app.InstallSandboxRun{}
	require.NoError(t, db.Where(app.InstallSandboxRun{ID: "run-1"}).First(&run).Error)
	require.Nil(t, run.CompositeError)
}
