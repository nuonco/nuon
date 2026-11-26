package helpers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

func stackVersionPostgres(t *testing.T) (*gorm.DB, *installhelpers.Helpers) {
	t.Helper()
	db := appliedHistoryPostgres(t)
	require.NoError(t, db.Exec("CREATE VIEW install_stack_versions_live_view_v1 AS "+viewsql.InstallStackVersionsLiveViewV1).Error)
	return db, installhelpers.New(installhelpers.Params{DB: db})
}

func newStackVersionStack(t require.TestingT, db *gorm.DB) *app.InstallStack {
	install := app.Install{ID: historyID("install", 0), CreatedByID: historyID("account", 0)}
	persistHistoryRow(t, db, &install)
	stack := app.InstallStack{ID: historyID("stack", 0), InstallID: install.ID,
		CreatedByID: historyID("account", 0), OrgID: historyID("org", 0)}
	persistHistoryRow(t, db, &stack)
	return &stack
}

func persistStackVersion(t require.TestingT, db *gorm.DB, n int, stack *app.InstallStack, cfgID string, status app.Status, createdAt, updatedAt time.Time) {
	version := app.InstallStackVersion{ID: historyID("stack-version", n), InstallID: stack.InstallID,
		InstallStackID: stack.ID, CreatedByID: historyID("account", 0), OrgID: historyID("org", 0),
		AppConfigID: cfgID, Status: app.CompositeStatus{Status: status}, CreatedAt: createdAt, UpdatedAt: updatedAt}
	persistHistoryRow(t, db, &version)
}

func setStackRef(t require.TestingT, db *gorm.DB, stack *app.InstallStack, ref app.AppConfigRef) {
	res := db.Model(&app.InstallStack{ID: stack.ID}).Update("app_config_ref", ref)
	require.NoError(t, res.Error)
	require.Equal(t, int64(1), res.RowsAffected)
}

func deleteStack(t require.TestingT, db *gorm.DB, stack *app.InstallStack) {
	res := db.Model(&app.InstallStack{ID: stack.ID}).Update("deleted_at", 1)
	require.NoError(t, res.Error)
	require.Equal(t, int64(1), res.RowsAffected)
}

func TestAppliedInstallStackVersionPostgres(t *testing.T) {
	t.Parallel()
	db, helpers := stackVersionPostgres(t)
	ctx := context.Background()
	at := func(minutes int) time.Time { return time.Date(2026, 1, 1, 0, minutes, 0, 0, time.UTC) }

	type stackVersionCase struct {
		seed    func(t *testing.T, db *gorm.DB, stack *app.InstallStack)
		wantID  string
		wantCfg string
	}
	tests := map[string]stackVersionCase{
		"applied pointer resolves that version over newer active ones": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusActive, at(0), at(0))
				persistStackVersion(t, db, 1, stack, historyID("config", 1), app.InstallStackVersionStatusActive, at(10), at(10))
				setStackRef(t, db, stack, app.AppConfigRef{
					ExpectedConfigID:    historyID("config", 1),
					AppliedConfigID:     historyID("config", 0),
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
			},
			wantID:  historyID("stack-version", 0),
			wantCfg: historyID("config", 0),
		},
		"a later write to an older version does not make it the applied one": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				// The older version was applied; a phone-home token mint then
				// touched it long after the newer version was created.
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusActive, at(0), at(30))
				persistStackVersion(t, db, 1, stack, historyID("config", 1), app.InstallStackVersionStatusActive, at(10), at(10))
				setStackRef(t, db, stack, app.AppConfigRef{})
			},
			wantID:  historyID("stack-version", 1),
			wantCfg: historyID("config", 1),
		},
		"legacy stack resolves the newest active version": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusActive, at(0), at(0))
				persistStackVersion(t, db, 1, stack, historyID("config", 1), app.InstallStackVersionStatusActive, at(10), at(10))
				setStackRef(t, db, stack, app.AppConfigRef{})
			},
			wantID:  historyID("stack-version", 1),
			wantCfg: historyID("config", 1),
		},
		"legacy stack without an active version resolves nothing": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusPendingUser, at(0), at(0))
				setStackRef(t, db, stack, app.AppConfigRef{})
			},
		},
		"a pointer without an applied config resolves nothing": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusActive, at(0), at(0))
				setStackRef(t, db, stack, app.AppConfigRef{
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
			},
		},
		"a soft-deleted stack resolves nothing": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistStackVersion(t, db, 0, stack, historyID("config", 0), app.InstallStackVersionStatusActive, at(0), at(0))
				deleteStack(t, db, stack)
			},
		},
		"stack without versions or pointer resolves nothing": {},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()

			stack := newStackVersionStack(t, tx)
			if test.seed != nil {
				test.seed(t, tx, stack)
			}

			got, err := helpers.WithDB(tx).AppliedInstallStackVersion(ctx, stack)
			require.NoError(t, err)
			if test.wantID == "" {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, test.wantID, got.ID)
			require.Equal(t, test.wantCfg, got.AppConfigID)
		})
	}
}
