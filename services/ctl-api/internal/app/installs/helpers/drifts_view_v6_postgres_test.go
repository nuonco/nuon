package helpers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

func driftsViewV6Postgres(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	db := appliedHistoryPostgres(t)
	require.NoError(t, db.AutoMigrate(&app.Component{}, &app.WorkflowStep{}))
	require.NoError(t, db.Exec("CREATE VIEW install_stack_versions_live_view_v1 AS "+viewsql.InstallStackVersionsLiveViewV1).Error)
	require.NoError(t, db.Exec("CREATE VIEW component_config_connections_selected_build_view_v1 AS "+viewsql.ComponentConfigConnectionsSelectedBuildViewV1).Error)
	name := views.DefaultViewName(db, &app.DriftedObject{}, 6)
	require.NoError(t, db.Exec("CREATE VIEW "+name+" AS "+viewsql.DriftsViewV6).Error)
	return db, name
}

type driftStackRow struct {
	ExpectedConfigID       string
	LiveID                 string
	ExpectedArtifactID     string
	ExpectedArtifactStatus string
	AttemptWorkflowID      string
	LiveConfigID           string
	LiveSource             string
	ExpectedState          string
	LiveState              string
	LiveOutdated           bool
}

func scanStackDriftRow(t require.TestingT, db *gorm.DB, name string) *driftStackRow {
	var rows []driftStackRow
	require.NoError(t, db.Table(name).
		Where("resource_type = ?", "install_stack").
		Scan(&rows).Error)
	if len(rows) == 0 {
		return nil
	}
	require.Len(t, rows, 1)
	return &rows[0]
}

func newDriftStack(t require.TestingT, db *gorm.DB, appConfigID string) *app.InstallStack {
	install := app.Install{ID: historyID("install", 0), CreatedByID: historyID("account", 0)}
	if appConfigID != "" {
		install.AppConfigID = appConfigID
	}
	persistHistoryRow(t, db, &install)
	stack := app.InstallStack{ID: historyID("stack", 0), InstallID: install.ID,
		CreatedByID: historyID("account", 0), OrgID: historyID("org", 0)}
	persistHistoryRow(t, db, &stack)
	return &stack
}

func persistDriftStackVersion(t require.TestingT, db *gorm.DB, n int, stack *app.InstallStack, cfgID string, status app.CompositeStatus, createdAt time.Time) {
	version := app.InstallStackVersion{ID: historyID("stack-version", n), InstallID: stack.InstallID,
		InstallStackID: stack.ID, CreatedByID: historyID("account", 0), OrgID: historyID("org", 0),
		AppConfigID: cfgID, Status: status, CreatedAt: createdAt, UpdatedAt: createdAt}
	persistHistoryRow(t, db, &version)
}

func setDriftStackRef(t require.TestingT, db *gorm.DB, stack *app.InstallStack, ref app.AppConfigRef) {
	res := db.Model(&app.InstallStack{ID: stack.ID}).Update("app_config_ref", ref)
	require.NoError(t, res.Error)
	require.Equal(t, int64(1), res.RowsAffected)
}

func persistDriftStackWorkflow(t require.TestingT, db *gorm.DB, n int, versionID string, planOnly bool, createdAt time.Time) {
	wf := app.Workflow{ID: historyID("workflow", n), CreatedByID: historyID("account", 0),
		OrgID: historyID("org", 0), OwnerID: historyID("workflow-owner", n), OwnerType: "installs",
		Type: app.WorkflowTypeReprovisionStack, CreatedAt: createdAt, UpdatedAt: createdAt, PlanOnly: planOnly}
	persistHistoryRow(t, db, &wf)
	step := app.WorkflowStep{ID: historyID("step", n), CreatedByID: historyID("account", 0),
		OrgID: historyID("org", 0), InstallWorkflowID: wf.ID,
		StepTargetType: string(app.WorkflowStepTargetTypeInstallStackVersions), StepTargetID: versionID,
		CreatedAt: createdAt, UpdatedAt: createdAt}
	persistHistoryRow(t, db, &step)
}

func TestDriftsViewV6StackArmPostgres(t *testing.T) {
	t.Parallel()
	db, viewName := driftsViewV6Postgres(t)
	at := func(minutes int) time.Time { return time.Date(2026, 1, 1, 0, minutes, 0, 0, time.UTC) }

	type stackArmCase struct {
		seed func(t *testing.T, db *gorm.DB, stack *app.InstallStack)
		want *driftStackRow
	}
	dupOf := func(versionID string) app.CompositeStatus {
		return app.CompositeStatus{Status: app.InstallStackVersionStatusOutdated,
			Metadata: map[string]any{app.InstallStackVersionDuplicateOfMetadataKey: versionID}}
	}
	tests := map[string]stackArmCase{
		"applied pointer resolves that version over newer active ones": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 1), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{
					ExpectedConfigID:    historyID("config", 1),
					AppliedConfigID:     historyID("config", 0),
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 0), LiveConfigID: historyID("config", 0),
				LiveSource: "recorded", ExpectedState: "not_in_config", LiveState: "deployed"},
		},
		"a later write to an older version does not make it the applied one": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				version := app.InstallStackVersion{ID: historyID("stack-version", 0), InstallID: stack.InstallID,
					InstallStackID: stack.ID, CreatedByID: historyID("account", 0), OrgID: historyID("org", 0),
					AppConfigID: historyID("config", 0), Status: app.CompositeStatus{Status: app.InstallStackVersionStatusActive},
					CreatedAt: at(0), UpdatedAt: at(30)}
				persistHistoryRow(t, db, &version)
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 1), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{})
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 1), LiveConfigID: historyID("config", 1),
				LiveSource: "inferred", ExpectedState: "not_in_config", LiveState: "deployed"},
		},
		"the newest version for the selected config is the desired one": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 1), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{
					AppliedConfigID:     historyID("config", 1),
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
				res := db.Model(&app.Install{ID: stack.InstallID}).Update("app_config_id", historyID("config", 1))
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 0), LiveConfigID: historyID("config", 1),
				LiveSource: "recorded", ExpectedArtifactID: historyID("stack-version", 0),
				ExpectedArtifactStatus: string(app.InstallStackVersionStatusActive), ExpectedConfigID: historyID("config", 1),
				ExpectedState: "ready", LiveState: "deployed", LiveOutdated: false},
		},
		"a newer template for the selected config is desired but not applied": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 1), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 1), app.CompositeStatus{Status: app.InstallStackVersionStatusPendingUser}, at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{
					AppliedConfigID:     historyID("config", 1),
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
				res := db.Model(&app.Install{ID: stack.InstallID}).Update("app_config_id", historyID("config", 1))
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 0), LiveConfigID: historyID("config", 1),
				LiveSource: "recorded", ExpectedArtifactID: historyID("stack-version", 1),
				ExpectedArtifactStatus: string(app.InstallStackVersionStatusPendingUser), ExpectedConfigID: historyID("config", 1),
				ExpectedState: "ready", LiveState: "deployed", LiveOutdated: true},
		},
		"an outdated newest version resolves to the active one it duplicates": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 0), dupOf(historyID("stack-version", 0)), at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{})
				res := db.Model(&app.Install{ID: stack.InstallID}).Update("app_config_id", historyID("config", 0))
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 0), LiveConfigID: historyID("config", 0),
				LiveSource: "inferred", ExpectedArtifactID: historyID("stack-version", 0),
				ExpectedArtifactStatus: string(app.InstallStackVersionStatusActive), ExpectedConfigID: historyID("config", 0),
				ExpectedState: "ready", LiveState: "deployed"},
		},
		"an outdated newest version with a missing duplicate resolves as itself": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), dupOf(historyID("stack-version", 9)), at(0))
				setDriftStackRef(t, db, stack, app.AppConfigRef{})
				res := db.Model(&app.Install{ID: stack.InstallID}).Update("app_config_id", historyID("config", 0))
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
			want: &driftStackRow{ExpectedArtifactID: historyID("stack-version", 0),
				ExpectedArtifactStatus: string(app.InstallStackVersionStatusOutdated), ExpectedConfigID: historyID("config", 0),
				ExpectedState: "ready", LiveState: "not_deployed"},
		},
		"the desired version's newest real workflow is reported over a plan-only one": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				persistDriftStackVersion(t, db, 1, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusPendingUser}, at(10))
				setDriftStackRef(t, db, stack, app.AppConfigRef{})
				persistDriftStackWorkflow(t, db, 0, historyID("stack-version", 1), true, at(20))
				persistDriftStackWorkflow(t, db, 1, historyID("stack-version", 1), false, at(15))
				res := db.Model(&app.Install{ID: stack.InstallID}).Update("app_config_id", historyID("config", 0))
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
			want: &driftStackRow{LiveID: historyID("stack-version", 0), LiveConfigID: historyID("config", 0),
				LiveSource: "inferred", ExpectedArtifactID: historyID("stack-version", 1),
				ExpectedArtifactStatus: string(app.InstallStackVersionStatusPendingUser), AttemptWorkflowID: historyID("workflow", 1),
				ExpectedConfigID: historyID("config", 0), ExpectedState: "ready", LiveState: "deployed",
				LiveOutdated: true},
		},
		"legacy stack without an active version has an unknown live state": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusPendingUser}, at(0))
				setDriftStackRef(t, db, stack, app.AppConfigRef{})
			},
			want: &driftStackRow{ExpectedState: "not_in_config", LiveState: "unknown"},
		},
		"a pointer without an applied config resolves nothing": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				setDriftStackRef(t, db, stack, app.AppConfigRef{
					AppliedConfigByType: app.AppConfigRefByTypeInstallStackVersions,
					AppliedConfigByID:   historyID("stack-version", 0),
				})
			},
			want: &driftStackRow{ExpectedState: "not_in_config", LiveState: "unknown"},
		},
		"a soft-deleted stack resolves nothing": {
			seed: func(t *testing.T, db *gorm.DB, stack *app.InstallStack) {
				persistDriftStackVersion(t, db, 0, stack, historyID("config", 0), app.CompositeStatus{Status: app.InstallStackVersionStatusActive}, at(0))
				res := db.Model(&app.InstallStack{ID: stack.ID}).Update("deleted_at", 1)
				require.NoError(t, res.Error)
				require.Equal(t, int64(1), res.RowsAffected)
			},
		},
		"stack without versions or pointer resolves nothing": {
			want: &driftStackRow{ExpectedState: "not_in_config", LiveState: "not_deployed"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()

			stack := newDriftStack(t, tx, "")
			if test.seed != nil {
				test.seed(t, tx, stack)
			}

			got := scanStackDriftRow(t, tx, viewName)
			if test.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *test.want, *got)
		})
	}
}
