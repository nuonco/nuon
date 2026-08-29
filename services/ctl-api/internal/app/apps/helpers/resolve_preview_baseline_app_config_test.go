package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestResolvePreviewBaselineAppConfig(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	trueVal := true

	t.Run("uses comparison base run on same branch", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRunWithConfig(t, db, "base-main", "branch-main", "cfg-main", string(app.AppBranchRunTypeManual), false, &trueVal, "success", now.Add(-time.Hour))
		insertBranchRunWithConfig(t, db, "base-seed", "branch-seed", "cfg-seed", string(app.AppBranchRunTypeGit), false, &trueVal, "success", now)
		insertBranchRunWithConfig(t, db, "preview-run", "branch-main", "", string(app.AppBranchRunTypeGitPreview), true, nil, "pending", now)
		require.NoError(t, db.Exec(`
			INSERT INTO app_branch_run_comparisons (
				id, org_id, created_by_id, created_at, updated_at, head_run_id, base_run_id
			) VALUES ('cmp-1', 'org-1', 'acc-1', ?, ?, 'preview-run', 'base-main')
		`, now, now).Error)

		h := &Helpers{db: db}
		out, err := h.ResolvePreviewBaselineAppConfig(ctx, "preview-run", "branch-main")
		require.NoError(t, err)
		require.Equal(t, "base-main", out.BaseRunID)
		require.Equal(t, "cfg-main", out.AppConfigID)
	})

	t.Run("falls back to find base run on branch", func(t *testing.T) {
		db := setupAppBranchRunComparisonDB(t)
		insertBranchRunWithConfig(t, db, "base-main", "branch-main", "cfg-main", string(app.AppBranchRunTypeManual), false, &trueVal, "success", now.Add(-time.Hour))
		insertBranchRunWithConfig(t, db, "base-seed", "branch-seed", "cfg-seed", string(app.AppBranchRunTypeGit), false, &trueVal, "success", now)

		h := &Helpers{db: db}
		out, err := h.ResolvePreviewBaselineAppConfig(ctx, "preview-run", "branch-main")
		require.NoError(t, err)
		require.Equal(t, "base-main", out.BaseRunID)
		require.Equal(t, "cfg-main", out.AppConfigID)
	})
}
