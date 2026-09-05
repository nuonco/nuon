package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestPkgWorkflowsFlowUpdateFlowPreflightErrors(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE install_workflows (id TEXT PRIMARY KEY, preflight_errors JSON, updated_at DATETIME, deleted_at INTEGER DEFAULT 0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO install_workflows (id, preflight_errors) VALUES (?, '[]')`, "iwf123").Error)

	warning := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "preflight.test",
		Severity: compositeerrors.SeverityWarning,
		Message:  "Preflight warning",
	}
	activity := &Activities{db: db}
	require.NoError(t, activity.PkgWorkflowsFlowUpdateFlowPreflightErrors(context.Background(), UpdateFlowPreflightErrorsRequest{
		FlowID:          "iwf123",
		PreflightErrors: []*compositeerrors.CompositeErrorData{warning},
	}))

	var workflow app.Workflow
	require.NoError(t, db.Select("id", "preflight_errors").Where(app.Workflow{ID: "iwf123"}).First(&workflow).Error)
	require.Len(t, workflow.PreflightErrors, 1)
	require.Equal(t, warning.Type, workflow.PreflightErrors[0].Type)
	require.Equal(t, warning.Severity, workflow.PreflightErrors[0].Severity)
	require.Equal(t, warning.Message, workflow.PreflightErrors[0].Message)

	require.NoError(t, activity.PkgWorkflowsFlowUpdateFlowPreflightErrors(context.Background(), UpdateFlowPreflightErrorsRequest{
		FlowID: "iwf123",
	}))
	require.NoError(t, db.Select("id", "preflight_errors").Where(app.Workflow{ID: "iwf123"}).First(&workflow).Error)
	require.Empty(t, workflow.PreflightErrors)
}
