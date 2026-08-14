package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func emptyWorkflowsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE install_workflows (id text, deleted_at integer default 0)").Error)
	return db
}

func requireNonRetryableNotFound(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var appErr *temporal.ApplicationError
	require.True(t, errors.As(err, &appErr))
	assert.True(t, appErr.NonRetryable())
	assert.Equal(t, "not found", appErr.Type())
}

func TestCheckWorkflowRetryableRecordNotFound(t *testing.T) {
	a := &Activities{db: emptyWorkflowsDB(t)}
	_, err := a.CheckWorkflowRetryable(context.Background(), CheckWorkflowRetryableRequest{WorkflowID: "wfl_missing"})
	requireNonRetryableNotFound(t, err)
}

func TestPkgWorkflowsFlowGetFlowRecordNotFound(t *testing.T) {
	a := &Activities{db: emptyWorkflowsDB(t)}
	_, err := a.PkgWorkflowsFlowGetFlow(context.Background(), GetFlowRequest{ID: "wfl_missing"})
	requireNonRetryableNotFound(t, err)
}
