package helpers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func ensureInstallActionsPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	db := appliedHistoryPostgres(t)
	require.NoError(t, db.AutoMigrate(&app.ActionWorkflowConfig{}, &app.InstallActionWorkflow{}))
	return db
}

func ensureInstallActionsContext() context.Context {
	ctx := cctx.SetOrgIDContext(context.Background(), "org-ensure-actions")
	return cctx.SetAccountIDContext(ctx, "acc-ensure-actions")
}

func installActionsByActionID(t *testing.T, db *gorm.DB, installID string) map[string]app.InstallActionWorkflow {
	t.Helper()
	var rows []app.InstallActionWorkflow
	require.NoError(t, db.Where(app.InstallActionWorkflow{InstallID: installID}).Find(&rows).Error)
	byAction := make(map[string]app.InstallActionWorkflow, len(rows))
	for _, row := range rows {
		byAction[row.ActionWorkflowID] = row
	}
	return byAction
}

// The install is still on its old config when this runs, so it must add the new
// config's actions without touching the install actions the old config uses.
func TestEnsureInstallActionsCreatesOnlyMissingActions(t *testing.T) {
	db := ensureInstallActionsPostgres(t)
	const installID = "inl-ensure-actions"

	for i, actionID := range []string{"acw-kept", "acw-added"} {
		persistHistoryRow(t, db, &app.ActionWorkflowConfig{
			ID:               historyID("awc-new", i),
			AppConfigID:      "cfg-new",
			ActionWorkflowID: actionID,
			CreatedByID:      "acc-ensure-actions",
			OrgID:            "org-ensure-actions",
		})
	}
	persistHistoryRow(t, db, &app.ActionWorkflowConfig{
		ID:               historyID("awc-old", 0),
		AppConfigID:      "cfg-old",
		ActionWorkflowID: "acw-old-only",
		CreatedByID:      "acc-ensure-actions",
		OrgID:            "org-ensure-actions",
	})

	for i, actionID := range []string{"acw-kept", "acw-old-only"} {
		persistHistoryRow(t, db, &app.InstallActionWorkflow{
			ID:               historyID("iaw", i),
			InstallID:        installID,
			ActionWorkflowID: actionID,
			StatusV2:         app.CompositeStatus{Status: app.StatusPending},
			CreatedByID:      "acc-ensure-actions",
			OrgID:            "org-ensure-actions",
		})
	}

	helpers := installhelpers.New(installhelpers.Params{DB: db})
	require.NoError(t, helpers.EnsureInstallActions(ensureInstallActionsContext(), installID, "cfg-new"))

	got := installActionsByActionID(t, db, installID)
	require.Len(t, got, 3)
	require.Equal(t, historyID("iaw", 0), got["acw-kept"].ID)
	require.Contains(t, got, "acw-added")
	require.Equal(t, app.StatusPending, got["acw-old-only"].StatusV2.Status)
	require.NotContains(t, got["acw-old-only"].StatusV2.Metadata, "removed_at_app_config_id")
}

// Step generation can be retried, so a second call must not add duplicates.
func TestEnsureInstallActionsIsIdempotent(t *testing.T) {
	db := ensureInstallActionsPostgres(t)
	const installID = "inl-ensure-actions"

	persistHistoryRow(t, db, &app.ActionWorkflowConfig{
		ID:               historyID("awc-new", 0),
		AppConfigID:      "cfg-new",
		ActionWorkflowID: "acw-added",
		CreatedByID:      "acc-ensure-actions",
		OrgID:            "org-ensure-actions",
	})

	helpers := installhelpers.New(installhelpers.Params{DB: db})
	for range 2 {
		require.NoError(t, helpers.EnsureInstallActions(ensureInstallActionsContext(), installID, "cfg-new"))
	}

	require.Len(t, installActionsByActionID(t, db, installID), 1)
}
