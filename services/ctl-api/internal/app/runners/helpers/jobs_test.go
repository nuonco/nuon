package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCreateJobExecutionResultIfAbsent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE runner_job_execution_results (
			id TEXT PRIMARY KEY,
			created_by_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at INTEGER NOT NULL DEFAULT 0,
			org_id TEXT,
			runner_job_execution_id TEXT NOT NULL,
			success BOOLEAN,
			error_code INTEGER,
			error_metadata TEXT,
			contents TEXT,
			contents_display BLOB,
			contents_gzip BLOB,
			contents_display_gzip BLOB,
			composite_error TEXT,
			UNIQUE (deleted_at, runner_job_execution_id)
		)
	`).Error)

	tests := map[string]struct {
		first    *app.RunnerJobExecutionResult
		fallback *app.RunnerJobExecutionResult
	}{
		"uncompressed": {
			first: &app.RunnerJobExecutionResult{
				ID:                   "uncompressed-first",
				CreatedByID:          "account-id",
				OrgID:                "org-id",
				RunnerJobExecutionID: "uncompressed-execution",
				ErrorCode:            42,
				Contents:             "specific handler result",
				ContentsDisplay:      []byte(`{"detail":"specific"}`),
			},
			fallback: &app.RunnerJobExecutionResult{
				ID:                   "uncompressed-fallback",
				CreatedByID:          "account-id",
				OrgID:                "org-id",
				RunnerJobExecutionID: "uncompressed-execution",
				Success:              true,
				Contents:             "generic fallback result",
				ContentsDisplay:      []byte(`{"detail":"generic"}`),
			},
		},
		"compressed": {
			first: &app.RunnerJobExecutionResult{
				ID:                   "compressed-first",
				CreatedByID:          "account-id",
				OrgID:                "org-id",
				RunnerJobExecutionID: "compressed-execution",
				ErrorCode:            42,
				ContentsGzip:         []byte("compressed-specific-result"),
				ContentsDisplayGzip:  []byte("compressed-specific-display"),
			},
			fallback: &app.RunnerJobExecutionResult{
				ID:                   "compressed-fallback",
				CreatedByID:          "account-id",
				OrgID:                "org-id",
				RunnerJobExecutionID: "compressed-execution",
				Success:              true,
				ContentsGzip:         []byte("compressed-generic-result"),
				ContentsDisplayGzip:  []byte("compressed-generic-display"),
			},
		},
	}

	ctx := context.Background()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			persisted, created, err := CreateJobExecutionResultIfAbsent(ctx, db, test.first)
			require.NoError(t, err)
			require.True(t, created)
			require.Equal(t, test.first.ID, persisted.ID)

			persisted, created, err = CreateJobExecutionResultIfAbsent(ctx, db, test.fallback)
			require.NoError(t, err)
			require.False(t, created)
			require.Equal(t, test.first.ID, persisted.ID)
			require.False(t, persisted.Success)
			require.Equal(t, test.first.ErrorCode, persisted.ErrorCode)
			require.Equal(t, test.first.Contents, persisted.Contents)
			require.Equal(t, test.first.ContentsDisplay, persisted.ContentsDisplay)
			require.Equal(t, test.first.ContentsGzip, persisted.ContentsGzip)
			require.Equal(t, test.first.ContentsDisplayGzip, persisted.ContentsDisplayGzip)
		})
	}

	var count int64
	require.NoError(t, db.Model(&app.RunnerJobExecutionResult{}).Count(&count).Error)
	require.EqualValues(t, len(tests), count)
}
