package helpers_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func appliedHistoryPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("NUON_APPLIED_HISTORY_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set NUON_APPLIED_HISTORY_POSTGRES_DSN to opt into a disposable loopback PostgreSQL database")
	}
	cfg, err := pgx.ParseConfig(dsn)
	require.NoError(t, err)
	hosts := []string{cfg.Host}
	for _, fallback := range cfg.Fallbacks {
		hosts = append(hosts, fallback.Host)
	}
	for _, host := range hosts {
		ip := net.ParseIP(host)
		require.True(t, host == "localhost" || ip != nil && ip.IsLoopback(), "refusing non-loopback PostgreSQL host %q", host)
	}
	cfg.ConnectTimeout = 5 * time.Second
	admin := sql.OpenDB(stdlib.GetConnector(*cfg))
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	name := "nuon_applied_history_" + strings.ToLower(rand.Text())
	_, err = admin.Exec("CREATE DATABASE " + pgx.Identifier{name}.Sanitize())
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP DATABASE " + pgx.Identifier{name}.Sanitize())
		require.NoError(t, err)
	})
	cfg.Database = name
	conn := sql.OpenDB(stdlib.GetConnector(*cfg))
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE EXTENSION hstore").Error)
	require.NoError(t, db.AutoMigrate(
		&app.AppConfig{}, &app.ComponentConfigConnection{}, &app.ComponentBuild{},
		&app.Install{}, &app.InstallComponent{}, &app.InstallDeploy{},
		&app.InstallSandbox{}, &app.InstallSandboxRun{}, &app.AppSandboxConfig{},
		&app.InstallStack{}, &app.InstallStackVersion{}, &app.Workflow{},
	))
	var version, database string
	require.NoError(t, db.Raw("SELECT version(), current_database()").Row().Scan(&version, &database))
	t.Logf("disposable PostgreSQL database=%s server=%s", database, version)
	return db
}

func historyID(kind string, n int) string {
	return fmt.Sprintf("%-20s%06d", kind, n)
}

func persistHistoryRow(t require.TestingT, db *gorm.DB, row any) {
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(row).Error)
}

type appliedReceiptFact struct {
	operationID string
	configID    string
	buildID     string
	consumed    map[string]string
}

type appliedHistory struct {
	db         *gorm.DB
	helpers    *installhelpers.Helpers
	activities *installactivities.Activities
	install    app.Install
	components []app.InstallComponent
	sandbox    app.InstallSandbox
	stack      app.InstallStack
	actual     []appliedReceiptFact
	refs       []app.AppConfigRef
	configs    []string
	fallback   string
	next       int
}

func newAppliedHistory(t require.TestingT, db *gorm.DB, n int) *appliedHistory {
	h := &appliedHistory{
		db:         db,
		helpers:    installhelpers.New(installhelpers.Params{DB: db}),
		activities: installactivities.New(installactivities.Params{DB: db}),
		install:    app.Install{ID: historyID("install", 0), CreatedByID: historyID("account", 0), AppConfigID: historyID("early-pin", 0)},
		actual:     make([]appliedReceiptFact, n+2),
		refs:       make([]app.AppConfigRef, n+2),
	}
	persistHistoryRow(t, db, &h.install)
	for i := range 3 {
		cfg := app.AppConfig{ID: historyID("config", i), BuildsPinned: true, BuildsLocked: true,
			CreatedByID: historyID("account", 0), OrgID: historyID("org", 0), StatusDescription: "active",
			SandboxBuildID: generics.NewNullString(historyID("sandbox-latest", i))}
		persistHistoryRow(t, db, &cfg)
		h.configs = append(h.configs, cfg.ID)
		persistHistoryRow(t, db, &app.AppSandboxConfig{ID: historyID("sandbox-config", i), AppConfigID: cfg.ID,
			CreatedByID: historyID("account", 0), OrgID: historyID("org", 0), AppID: historyID("app", 0)})
		for j := range n {
			persistHistoryRow(t, db, &app.ComponentConfigConnection{
				ID: historyID("connection", i*n+j), AppConfigID: cfg.ID, ComponentID: historyID("component", j),
				CreatedByID: historyID("account", 0),
			})
		}
	}
	for i := range n {
		component := app.InstallComponent{ID: historyID("install-component", i), InstallID: h.install.ID, ComponentID: historyID("component", i)}
		persistHistoryRow(t, db, &component)
		h.components = append(h.components, component)
	}
	h.sandbox = app.InstallSandbox{ID: historyID("sandbox", 0), InstallID: h.install.ID,
		CreatedByID: historyID("account", 0), Status: app.InstallSandboxStatusEmpty, StatusDescription: "empty"}
	h.stack = app.InstallStack{ID: historyID("stack", 0), InstallID: h.install.ID,
		CreatedByID: historyID("account", 0), OrgID: historyID("org", 0)}
	persistHistoryRow(t, db, &h.sandbox)
	persistHistoryRow(t, db, &h.stack)
	return h
}

func (h *appliedHistory) append(t require.TestingT, resource, revision int, event string, legacy bool) {
	h.next++
	at := time.Date(2026, 1, 1, 0, 0, h.next, 0, time.UTC)
	cfgID := h.configs[revision]
	id := historyID("operation", h.next)
	buildID := historyID("build", h.next)
	wf := app.Workflow{ID: historyID("workflow", h.next), PlanOnly: event == "plan",
		CreatedByID: historyID("account", 0), Type: app.WorkflowTypeManualDeploy}
	persistHistoryRow(t, h.db, &wf)
	var appliedAt *time.Time
	if event == "apply" || event == "health-failed" || event == "clear" || event == "recover" {
		appliedAt = &at
	}
	advances := event == "apply" || event == "health-failed" || event == "sync" || event == "clear"
	if resource < len(h.components) {
		connID := historyID("connection", revision*len(h.components)+resource)
		persistHistoryRow(t, h.db, &app.ComponentBuild{ID: buildID, ComponentConfigConnectionID: connID,
			CreatedByID: historyID("account", 0), Status: app.ComponentBuildStatusActive})
		kind := app.InstallDeployTypeApply
		status := app.InstallDeployStatusActive
		switch event {
		case "recover":
			kind = app.InstallDeployTypeRecover
		case "clear":
			kind = app.InstallDeployTypeTeardown
		case "sync":
			kind = app.InstallDeployTypeSync
		case "fail":
			status = app.InstallDeployStatusError
		case "health-failed":
			status = app.InstallDeployStatusHealthFailed
		}
		consumed := map[string]string{}
		for i := range h.components {
			if i != resource && h.actual[i].buildID != "" {
				consumed[h.components[i].ComponentID] = h.actual[i].buildID
			}
		}
		deploy := app.InstallDeploy{ID: id, InstallComponentID: h.components[resource].ID,
			ComponentBuildID: buildID, AppConfigID: cfgID, DependencyBuildIDs: consumed,
			CreatedByID: historyID("account", 0), StatusDescription: "attempt",
			Type: kind, Status: status, AppliedAt: appliedAt, PlannedAt: &at,
			CreatedAt: at, UpdatedAt: at, InstallWorkflowID: &wf.ID}
		persistHistoryRow(t, h.db, &deploy)
		if advances && !legacy || event == "recover" {
			require.NoError(t, h.activities.SetDeployAppliedAt(context.Background(), installactivities.SetDeployAppliedAtRequest{DeployID: id}))
		}
		if advances {
			h.actual[resource] = appliedReceiptFact{operationID: id, configID: cfgID, buildID: buildID, consumed: consumed}
		}
	} else if resource == len(h.components) {
		kind := app.SandboxRunTypeReprovision
		if event == "clear" {
			kind = app.SandboxRunTypeDeprovision
		}
		run := app.InstallSandboxRun{ID: id, InstallID: h.install.ID, InstallSandboxID: &h.sandbox.ID,
			AppSandboxConfigID: historyID("sandbox-config", revision), AppSandboxBuildID: buildID,
			CreatedByID: historyID("account", 0),
			RunType:     kind, AppliedAt: appliedAt, PlannedAt: &at, CreatedAt: at, UpdatedAt: at, InstallWorkflowID: &wf.ID}
		persistHistoryRow(t, h.db, &run)
		if advances && !legacy {
			require.NoError(t, h.activities.SetSandboxRunAppliedAt(context.Background(), installactivities.SetSandboxRunAppliedAtRequest{SandboxRunID: id}))
		}
		if advances {
			h.actual[resource] = appliedReceiptFact{operationID: id, configID: cfgID, buildID: buildID}
		}
	} else {
		version := app.InstallStackVersion{ID: id, InstallID: h.install.ID, InstallStackID: h.stack.ID,
			CreatedByID: historyID("account", 0), OrgID: historyID("org", 0),
			AppConfigID: cfgID, CreatedAt: at, Status: app.CompositeStatus{Status: app.InstallStackVersionStatusActive}}
		persistHistoryRow(t, h.db, &version)
		require.NoError(t, h.helpers.RecordInstallStackVersionApplied(context.Background(), &version, installhelpers.PhoneHomeRequestTypeUpdate))
		h.actual[resource] = appliedReceiptFact{operationID: id, configID: cfgID}
	}
	if advances && event == "clear" {
		h.actual[resource] = appliedReceiptFact{}
	}
	if !legacy && advances {
		if event == "clear" {
			h.refs[resource] = app.AppConfigRef{ExpectedConfigID: h.refs[resource].ExpectedConfigID}
		} else {
			kind := app.AppConfigRefByTypeInstallDeploys
			if resource == len(h.components) {
				kind = app.AppConfigRefByTypeInstallSandboxRuns
			}
			h.refs[resource] = app.AppConfigRef{ExpectedConfigID: cfgID, AppliedConfigID: cfgID,
				AppliedConfigByID: id, AppliedConfigByType: kind, AppliedConfigAt: &at}
			if resource == len(h.components)+1 {
				h.refs[resource].ExpectedConfigID = ""
				h.refs[resource].AppliedConfigByType = app.AppConfigRefByTypeInstallStackVersions
				h.refs[resource].AppliedConfigAt = nil
			}
		}
	}
}

func (h *appliedHistory) check(t require.TestingT) {
	ctx := context.Background()
	components := make([]app.InstallComponent, 0, len(h.components))
	require.NoError(t, h.db.Where(app.InstallComponent{InstallID: h.install.ID}).Order("component_id").Find(&components).Error)
	var sandbox app.InstallSandbox
	require.NoError(t, h.db.Where(app.InstallSandbox{ID: h.sandbox.ID}).First(&sandbox).Error)
	deploys, err := h.helpers.AppliedInstallDeploys(ctx, components)
	require.NoError(t, err)
	run, err := h.helpers.AppliedInstallSandboxRun(ctx, &sandbox)
	require.NoError(t, err)
	baseline, err := h.helpers.InstallCompositeBaselines(ctx, &h.install)
	require.NoError(t, err)
	require.Equal(t, h.fallback, baseline.Fallback)
	require.Len(t, baseline.Components, len(h.components))
	for i := range components {
		component := &components[i]
		if component.AppConfigRef.AppliedConfigAt != nil {
			component.AppConfigRef.AppliedConfigAt = generics.ToPtr(component.AppConfigRef.AppliedConfigAt.UTC())
		}
		require.Equal(t, h.refs[i], component.AppConfigRef)
		want := h.actual[i]
		require.Equal(t, want.configID, baseline.Components[component.ComponentID], "component %d at step %d", i, h.next)
		if want.operationID == "" {
			require.NotContains(t, deploys, component.ID)
			require.Nil(t, baseline.ComponentBuilds[component.ComponentID])
			continue
		}
		require.Contains(t, deploys, component.ID)
		require.Equal(t, want.operationID, deploys[component.ID].ID)
		require.Equal(t, want.buildID, baseline.ComponentBuilds[component.ComponentID].ID)
		require.Equal(t, want.consumed, baseline.DependencyBuildIDs[component.ComponentID])
	}
	want := h.actual[len(h.components)]
	if sandbox.AppConfigRef.AppliedConfigAt != nil {
		sandbox.AppConfigRef.AppliedConfigAt = generics.ToPtr(sandbox.AppConfigRef.AppliedConfigAt.UTC())
	}
	require.Equal(t, h.refs[len(h.components)], sandbox.AppConfigRef)
	require.Equal(t, want.configID, baseline.Sandbox)
	require.NotNil(t, baseline.SandboxBuildID)
	require.Equal(t, want.buildID, *baseline.SandboxBuildID)
	if want.operationID == "" {
		require.Nil(t, run)
	} else {
		require.NotNil(t, run)
		require.Equal(t, want.operationID, run.ID)
	}
	var stack app.InstallStack
	require.NoError(t, h.db.Where(app.InstallStack{ID: h.stack.ID}).First(&stack).Error)
	if h.refs[len(h.components)+1].AppliedConfigID != "" {
		require.NotNil(t, stack.AppConfigRef.AppliedConfigAt)
		stack.AppConfigRef.AppliedConfigAt = nil
	}
	require.Equal(t, h.refs[len(h.components)+1], stack.AppConfigRef)
	require.Equal(t, h.actual[len(h.components)+1].configID, baseline.Stack)
}

func TestAppliedHistoryPostgresProperties(t *testing.T) {
	t.Parallel()
	db := appliedHistoryPostgres(t)
	hegel.Test(t, func(t *hegel.T) {
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		n := hegel.Draw(t, hegel.Integers(2, 4))
		h := newAppliedHistory(t, tx, n)
		h.check(t)
		for i := range n + 2 {
			h.append(t, i, hegel.Draw(t, hegel.Integers(0, 2)), "apply", false)
			h.check(t)
		}
		for range hegel.Draw(t, hegel.Integers(5, 16)) {
			resource := hegel.Draw(t, hegel.Integers(0, n+1))
			event := "apply"
			if resource < n {
				event = hegel.Draw(t, hegel.SampledFrom([]string{"apply", "fail", "plan", "recover", "clear", "health-failed"}))
			} else if resource == n {
				event = hegel.Draw(t, hegel.SampledFrom([]string{"apply", "fail", "plan", "clear"}))
			}
			h.append(t, resource, hegel.Draw(t, hegel.Integers(0, 2)), event, false)
			h.check(t)
		}
	}, hegel.WithTestCases(60))
}

func TestAppliedLegacyHistoryPostgresProperties(t *testing.T) {
	t.Parallel()
	db := appliedHistoryPostgres(t)
	hegel.Test(t, func(t *hegel.T) {
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		h := newAppliedHistory(t, tx, 2)
		for i := range 3 {
			event := "apply"
			if i < 2 && hegel.Draw(t, hegel.Booleans()) {
				event = "sync"
			}
			h.append(t, i, hegel.Draw(t, hegel.Integers(0, 2)), event, true)
			h.check(t)
		}
		for range hegel.Draw(t, hegel.Integers(4, 12)) {
			resource := hegel.Draw(t, hegel.Integers(0, 2))
			events := []string{"apply", "fail", "plan", "clear"}
			if resource < 2 {
				events = append(events, "sync", "recover", "health-failed")
			}
			h.append(t, resource, hegel.Draw(t, hegel.Integers(0, 2)), hegel.Draw(t, hegel.SampledFrom(events)), true)
			h.check(t)
		}
	}, hegel.WithTestCases(60))
}

func TestAppliedHistoryPostgresBoundaries(t *testing.T) {
	t.Parallel()
	db := appliedHistoryPostgres(t)
	tests := map[string]func(*testing.T, *appliedHistory){
		"later attempts do not advance pinned or legacy state": func(t *testing.T, h *appliedHistory) {
			for _, resource := range []int{0, 1, 2} {
				legacy := resource == 1
				h.append(t, resource, 0, "apply", legacy)
				for _, event := range []string{"fail", "plan", "unapplied"} {
					h.append(t, resource, 1, event, legacy)
					h.check(t)
				}
				if resource < 2 {
					h.append(t, resource, 2, "recover", legacy)
				}
			}
		},
		"pinned receipts beat newer applied rows": func(t *testing.T, h *appliedHistory) {
			for _, resource := range []int{0, 2} {
				h.append(t, resource, 0, "apply", false)
				pinned := h.actual[resource]
				h.append(t, resource, 1, "apply", true)
				h.actual[resource] = pinned
			}
		},
		"deleted historical configs remain available": func(t *testing.T, h *appliedHistory) {
			h.append(t, 0, 0, "apply", false)
			h.append(t, 1, 1, "apply", false)
			h.append(t, 2, 0, "apply", false)
			h.append(t, 3, 2, "apply", false)
			require.NoError(t, h.db.Delete(&app.ComponentConfigConnection{}, app.ComponentConfigConnection{ID: historyID("connection", 0)}).Error)
			require.NoError(t, h.db.Delete(&app.AppSandboxConfig{}, app.AppSandboxConfig{ID: historyID("sandbox-config", 0)}).Error)
			require.NoError(t, h.db.Delete(&app.AppConfig{}, app.AppConfig{ID: h.configs[0]}).Error)
		},
		"sandbox receipt can be recorded after historical config deletion": func(t *testing.T, h *appliedHistory) {
			h.append(t, 2, 0, "apply", true)
			require.NoError(t, h.db.Delete(&app.AppSandboxConfig{}, app.AppSandboxConfig{ID: historyID("sandbox-config", 0)}).Error)
			var run app.InstallSandboxRun
			require.NoError(t, h.db.Where(app.InstallSandboxRun{ID: h.actual[2].operationID}).First(&run).Error)
			require.NotNil(t, run.AppliedAt)
			at := run.AppliedAt.UTC()
			require.NoError(t, h.activities.SetSandboxRunAppliedAt(context.Background(), installactivities.SetSandboxRunAppliedAtRequest{SandboxRunID: run.ID}))
			h.refs[2] = app.AppConfigRef{ExpectedConfigID: h.configs[0], AppliedConfigID: h.configs[0],
				AppliedConfigByType: app.AppConfigRefByTypeInstallSandboxRuns, AppliedConfigByID: h.actual[2].operationID, AppliedConfigAt: &at}
		},
		"frozen consumed receipts survive newer dependency builds": func(t *testing.T, h *appliedHistory) {
			h.append(t, 0, 0, "apply", false)
			h.append(t, 1, 0, "apply", false)
			h.append(t, 2, 0, "apply", false)
			h.append(t, 0, 0, "apply", false)
			require.NotEqual(t, h.actual[0].buildID, h.actual[1].consumed[h.components[0].ComponentID])
			require.NoError(t, h.db.Model(&app.ComponentConfigConnection{}).
				Where(app.ComponentConfigConnection{AppConfigID: h.configs[0]}).
				Update("latest_build_id", h.actual[0].buildID).Error)
			require.NoError(t, h.db.Model(&app.AppConfig{}).Where(app.AppConfig{ID: h.configs[0]}).
				Update("sandbox_build_id", historyID("new-sandbox-build", 0)).Error)
		},
		"explicit empty markers do not rediscover historical applied rows": func(t *testing.T, h *appliedHistory) {
			for _, resource := range []int{0, 2} {
				h.append(t, resource, 0, "apply", false)
				kind := app.AppConfigRefByTypeInstallDeploys
				var model any = &app.InstallComponent{ID: h.components[0].ID}
				if resource == 2 {
					kind = app.AppConfigRefByTypeInstallSandboxRuns
					model = &app.InstallSandbox{ID: h.sandbox.ID}
				}
				ref := app.AppConfigRef{ExpectedConfigID: h.configs[1], AppliedConfigByType: kind,
					AppliedConfigByID: historyID("explicit-clear", resource)}
				require.NoError(t, h.db.Model(model).Update("app_config_ref", ref).Error)
				h.refs[resource], h.actual[resource] = ref, appliedReceiptFact{}
			}
		},
		"never applied resources do not inherit the install applied config": func(t *testing.T, h *appliedHistory) {
			h.fallback = h.configs[2]
			h.install.AppConfigRef.AppliedConfigID = h.fallback
			h.actual[3].configID = h.fallback
			require.NoError(t, h.db.Model(&app.Install{ID: h.install.ID}).Update("app_config_ref", h.install.AppConfigRef).Error)
		},
		"teardown and deprovision clear actual but preserve expected": func(t *testing.T, h *appliedHistory) {
			for _, resource := range []int{0, 2} {
				h.append(t, resource, 0, "apply", false)
				h.append(t, resource, 1, "clear", false)
				h.append(t, resource, 2, "fail", false)
				require.Equal(t, h.configs[0], h.refs[resource].ExpectedConfigID)
			}
		},
		"legacy applied time beats attempt creation order": func(t *testing.T, h *appliedHistory) {
			for _, resource := range []int{0, 2} {
				h.append(t, resource, 0, "apply", true)
				first := h.actual[resource]
				h.append(t, resource, 1, "apply", true)
				var model any = &app.InstallDeploy{ID: first.operationID}
				if resource == 2 {
					model = &app.InstallSandboxRun{ID: first.operationID}
				}
				require.NoError(t, h.db.Model(model).Update("applied_at", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)).Error)
				h.actual[resource] = first
			}
		},
		"legacy sync respects v2 status over legacy status": func(t *testing.T, h *appliedHistory) {
			h.append(t, 0, 0, "sync", true)
			first := h.actual[0]
			h.append(t, 0, 1, "sync", true)
			require.NoError(t, h.db.Model(&app.InstallDeploy{ID: h.actual[0].operationID}).
				Update("status_v2", app.CompositeStatus{Status: app.StatusError}).Error)
			h.actual[0] = first
		},
		"legacy sync accepts v2 active despite legacy error": func(t *testing.T, h *appliedHistory) {
			h.append(t, 0, 0, "apply", true)
			h.append(t, 0, 1, "sync", true)
			require.NoError(t, h.db.Model(&app.InstallDeploy{ID: h.actual[0].operationID}).Updates(map[string]any{
				"status": app.InstallDeployStatusError, "status_v2": app.CompositeStatus{Status: app.Status(app.InstallDeployStatusActive)},
			}).Error)
		},
		"plan-only active legacy sync is not applied": func(t *testing.T, h *appliedHistory) {
			h.append(t, 0, 0, "sync", true)
			first := h.actual[0]
			h.append(t, 0, 1, "sync", true)
			require.NoError(t, h.db.Model(&app.Workflow{ID: historyID("workflow", h.next)}).Update("plan_only", true).Error)
			h.actual[0] = first
		},
		"legacy active stack version supplies missing stack reference": func(t *testing.T, h *appliedHistory) {
			h.append(t, 3, 0, "apply", false)
			require.NoError(t, h.db.Model(&app.InstallStack{ID: h.stack.ID}).Update("app_config_ref", app.AppConfigRef{}).Error)
			h.refs[3] = app.AppConfigRef{}
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			h := newAppliedHistory(t, tx, 2)
			test(t, h)
			h.check(t)
		})
	}
}

func TestAppliedSandboxBuildReceiptPostgresProperties(t *testing.T) {
	t.Parallel()
	db := appliedHistoryPostgres(t)
	hegel.Test(t, func(t *hegel.T) {
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()
		h := newAppliedHistory(t, tx, 2)
		pinned, locked := hegel.Draw(t, hegel.Booleans()), hegel.Draw(t, hegel.Booleans())
		receipt := hegel.Draw(t, hegel.Booleans())
		h.append(t, 2, 0, "apply", hegel.Draw(t, hegel.Booleans()))
		require.NoError(t, tx.Model(&app.AppConfig{ID: h.configs[0]}).
			Updates(map[string]any{"builds_pinned": pinned, "builds_locked": locked}).Error)
		if !receipt {
			require.NoError(t, tx.Model(&app.InstallSandboxRun{ID: h.actual[2].operationID}).Update("app_sandbox_build_id", "").Error)
			h.actual[2].buildID = ""
			if pinned && locked {
				h.actual[2].buildID = historyID("sandbox-latest", 0)
			}
		}
		h.check(t)
	}, hegel.WithTestCases(32))
}

func TestRecordAppliedConfigPrecedencePostgres(t *testing.T) {
	db := appliedHistoryPostgres(t)
	tests := map[string]struct {
		deployConfig string
		flowType     app.WorkflowType
		flowConfig   string
		wantConfig   string
	}{
		"explicit deploy config beats workflow and build":  {historyID("config", 2), app.WorkflowTypeAppBranchConfigUpdate, historyID("config", 1), historyID("config", 2)},
		"branch workflow target beats reused build config": {"", app.WorkflowTypeAppBranchConfigUpdate, historyID("config", 1), historyID("config", 1)},
		"unrelated workflow metadata is not a config pin":  {"", app.WorkflowTypeManualDeploy, historyID("config", 1), historyID("config", 0)},
		"legacy build config remains the final fallback":   {"", app.WorkflowTypeAppBranchConfigUpdate, "", historyID("config", 0)},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			h := newAppliedHistory(t, tx, 2)
			h.append(t, 0, 0, "apply", true)
			operationID := h.actual[0].operationID
			require.NoError(t, tx.Delete(&app.ComponentConfigConnection{}, app.ComponentConfigConnection{ID: historyID("connection", 0)}).Error)
			require.NoError(t, tx.Model(&app.InstallDeploy{ID: operationID}).Update("app_config_id", test.deployConfig).Error)
			require.NoError(t, tx.Model(&app.Workflow{ID: historyID("workflow", h.next)}).Updates(map[string]any{
				"type": test.flowType, "metadata": pgtype.Hstore{"new_app_config_id": &test.flowConfig},
			}).Error)
			var deploy app.InstallDeploy
			require.NoError(t, tx.Where(app.InstallDeploy{ID: operationID}).First(&deploy).Error)
			require.NotNil(t, deploy.AppliedAt)
			at := deploy.AppliedAt.UTC()
			require.NoError(t, h.activities.SetDeployAppliedAt(context.Background(), installactivities.SetDeployAppliedAtRequest{DeployID: operationID}))
			h.actual[0].configID = test.wantConfig
			h.refs[0] = app.AppConfigRef{ExpectedConfigID: test.wantConfig, AppliedConfigID: test.wantConfig,
				AppliedConfigByType: app.AppConfigRefByTypeInstallDeploys, AppliedConfigByID: operationID, AppliedConfigAt: &at}
			h.check(t)
		})
	}
}
