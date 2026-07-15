package controlplanejob

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExecutionStatusForError(t *testing.T) {
	tests := map[string]struct {
		err    error
		status app.RunnerJobExecutionStatus
	}{
		"failure": {
			err:    errors.New("build failed"),
			status: app.RunnerJobExecutionStatusFailed,
		},
		"timeout": {
			err:    temporal.NewTimeoutError(enumspb.TIMEOUT_TYPE_START_TO_CLOSE, errors.New("deadline exceeded")),
			status: app.RunnerJobExecutionStatusTimedOut,
		},
		"cancellation": {
			err:    temporal.NewCanceledError(),
			status: app.RunnerJobExecutionStatusCancelled,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.status, executionStatusForError(test.err))
		})
	}
}

func TestApplyComponentBuildSourceIdentity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE component_builds (
			id TEXT PRIMARY KEY,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0,
			source_ref TEXT,
			source_image TEXT,
			resolved_tag TEXT,
			source_digest TEXT,
			source_media_type TEXT,
			resolved_at DATETIME,
			no_op BOOLEAN
		)
	`).Error)
	require.NoError(t, db.Exec("INSERT INTO component_builds (id) VALUES (?)", "build-id").Error)

	activities := &Activities{db: db, l: zap.NewNop()}
	resolvedAt := time.Date(2026, time.July, 13, 20, 15, 30, 123, time.UTC)
	err = activities.applyComponentBuildSourceIdentity(context.Background(), &app.RunnerJob{
		OwnerType: "component_builds",
		OwnerID:   "build-id",
	}, &models.ServiceCreateRunnerJobExecutionResultRequest{
		SourceRef:       "example.com/app:0.8.0",
		SourceImage:     "example.com/app",
		ResolvedTag:     "0.8.0",
		SourceDigest:    "sha256:abc123",
		SourceMediaType: "application/vnd.oci.image.manifest.v1+json",
		ResolvedAt:      resolvedAt.Format(time.RFC3339Nano),
		NoOp:            true,
	})
	require.NoError(t, err)

	var build app.ComponentBuild
	require.NoError(t, db.First(&build, "id = ?", "build-id").Error)
	require.Equal(t, "example.com/app:0.8.0", build.SourceRef)
	require.Equal(t, "example.com/app", build.SourceImage)
	require.Equal(t, "0.8.0", build.ResolvedTag)
	require.Equal(t, "sha256:abc123", build.SourceDigest)
	require.Equal(t, "application/vnd.oci.image.manifest.v1+json", build.SourceMediaType)
	require.WithinDuration(t, resolvedAt, *build.ResolvedAt, time.Nanosecond)
	require.True(t, build.NoOp)
}
