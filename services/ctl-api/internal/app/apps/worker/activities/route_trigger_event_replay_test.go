package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func setupReplayRoutingDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	statements := []string{
		`CREATE TABLE orgs (id TEXT PRIMARY KEY, features TEXT, deleted_at INTEGER DEFAULT 0)`,
		`CREATE TABLE triggers (
			id TEXT PRIMARY KEY, org_id TEXT, name TEXT, created_by_id TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at INTEGER DEFAULT 0
		)`,
		`CREATE TABLE trigger_events (
			id TEXT PRIMARY KEY, created_at DATETIME, updated_at DATETIME, deleted_at INTEGER DEFAULT 0,
			trigger_id TEXT, org_id TEXT, external_id TEXT, dedupe_id TEXT, source TEXT, event_type TEXT,
			received_at DATETIME, payload TEXT, headers TEXT,
			routing_status TEXT, routing_generation_token TEXT, routing_error TEXT,
			routing_started_at DATETIME, routing_completed_at DATETIME,
			match_count INTEGER DEFAULT 0, waiter_match_count INTEGER DEFAULT 0, dispatch_count INTEGER DEFAULT 0,
			match_explanations TEXT, explanations_truncated INTEGER DEFAULT 0
		)`,
		`CREATE TABLE app_configs (
			id TEXT PRIMARY KEY, created_at DATETIME, updated_at DATETIME, deleted_at INTEGER DEFAULT 0,
			org_id TEXT, app_id TEXT, status TEXT, labels TEXT
		)`,
		`CREATE TABLE trigger_rules (
			id TEXT PRIMARY KEY, created_at DATETIME, updated_at DATETIME, deleted_at INTEGER DEFAULT 0,
			org_id TEXT, app_id TEXT, app_config_id TEXT, trigger_id TEXT, name TEXT,
			enabled INTEGER DEFAULT 0, suspended_at DATETIME, valid_from DATETIME, valid_to DATETIME,
			event_types TEXT, filters TEXT, target_type TEXT, app_branch_id TEXT, runbook_id TEXT,
			install_name TEXT, input_mappings TEXT, force INTEGER DEFAULT 0, plan_only INTEGER DEFAULT 0,
			config_hash TEXT
		)`,
		`CREATE TABLE event_dispatches (
			id TEXT PRIMARY KEY, created_by_id TEXT, created_at DATETIME, updated_at DATETIME,
			org_id TEXT, app_id TEXT, trigger_event_id TEXT, trigger_rule_id TEXT, replay_id TEXT,
			idempotency_key TEXT UNIQUE, target_type TEXT, target_id TEXT, runbook_config_id TEXT,
			mapped_inputs TEXT, status TEXT, attempts INTEGER DEFAULT 0,
			generation_token TEXT, execution_token TEXT, next_attempt_at DATETIME, error TEXT,
			queue_signal_id TEXT, result_resource_type TEXT, result_resource_id TEXT, workflow_id TEXT,
			started_at DATETIME, triggered_at DATETIME, failed_at DATETIME
		)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
	return db
}

func TestReplayRoutingDoesNotMutateEventLedger(t *testing.T) {
	db := setupReplayRoutingDB(t)
	a := &Activities{db: db}
	ctx := context.Background()

	const (
		orgID     = "orgr0000000000000000000000"
		triggerID = "trg00000000000000000000000"
		eventID   = "eve00000000000000000000000"
		configID  = "apc00000000000000000000000"
		ruleID    = "trl00000000000000000000000"
	)

	require.NoError(t, db.Exec(`INSERT INTO orgs (id, features) VALUES (?, ?)`, orgID, `{"triggers":true}`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO triggers (id, org_id, name, created_by_id, created_at, updated_at)
		VALUES (?, ?, 'ci-trigger', 'acct000000000000000000000', datetime('now'), datetime('now'))
	`, triggerID, orgID).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO trigger_events (
			id, created_at, updated_at, trigger_id, org_id, external_id, dedupe_id, source, event_type,
			received_at, payload, routing_status, routing_generation_token, routing_error,
			routing_completed_at, match_count, waiter_match_count, dispatch_count,
			match_explanations, explanations_truncated
		) VALUES (
			?, datetime('now'), datetime('now'), ?, ?, 'ext-1', 'ext-1', 'github', 'push',
			datetime('now'), '{"attempt":1}', 'matched', 'generation-original', '',
			datetime('now'), 2, 1, 2,
			'[{"rule_id":"original"}]', 1
		)
	`, eventID, triggerID, orgID).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO app_configs (id, created_at, updated_at, org_id, app_id, status)
		VALUES (?, datetime('now'), datetime('now'), ?, 'app00000000000000000000000', 'active')
	`, configID, orgID).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO trigger_rules (
			id, created_at, updated_at, org_id, app_id, app_config_id, trigger_id, name, enabled,
			valid_from, target_type, runbook_id, install_name, config_hash
		) VALUES (
			?, datetime('now'), datetime('now'), ?, 'app00000000000000000000000', ?, ?, 'deploy-rule', 1,
			datetime('now', '-1 hour'), 'runbook', NULL, 'primary', 'hash'
		)
	`, ruleID, orgID, configID, triggerID).Error)

	var before app.TriggerEvent
	require.NoError(t, db.Where(app.TriggerEvent{ID: eventID}).First(&before).Error)

	resp, err := a.routeTriggerEventReplay(ctx, RouteTriggerEventRequest{EventID: eventID, ReplayID: "replay-1"})
	require.NoError(t, err)
	require.Empty(t, resp.Dispatches)

	var after app.TriggerEvent
	require.NoError(t, db.Where(app.TriggerEvent{ID: eventID}).First(&after).Error)
	require.Equal(t, before.RoutingStatus, after.RoutingStatus)
	require.Equal(t, before.RoutingError, after.RoutingError)
	require.Equal(t, before.MatchCount, after.MatchCount)
	require.Equal(t, before.WaiterMatchCount, after.WaiterMatchCount)
	require.Equal(t, before.DispatchCount, after.DispatchCount)
	require.Equal(t, before.MatchExplanations, after.MatchExplanations)
	require.Equal(t, before.ExplanationsTruncated, after.ExplanationsTruncated)
	require.NotNil(t, after.RoutingGenerationToken)
	require.Equal(t, "generation-original", *after.RoutingGenerationToken)
	require.Equal(t, before.RoutingCompletedAt.UTC(), after.RoutingCompletedAt.UTC())

	var dispatches []app.EventDispatch
	require.NoError(t, db.Where(app.EventDispatch{TriggerEventID: eventID}).Find(&dispatches).Error)
	require.Len(t, dispatches, 1)
	require.NotNil(t, dispatches[0].ReplayID)
	require.Equal(t, "replay-1", *dispatches[0].ReplayID)
	require.Equal(t, app.EventDispatchStatusDeadLettered, dispatches[0].Status)
	require.Equal(t, orgID, dispatches[0].OrgID)

	resp, err = a.routeTriggerEventReplay(ctx, RouteTriggerEventRequest{EventID: eventID, ReplayID: "replay-1"})
	require.NoError(t, err)
	require.Empty(t, resp.Dispatches)
	require.NoError(t, db.Where(app.EventDispatch{TriggerEventID: eventID}).Find(&dispatches).Error)
	require.Len(t, dispatches, 1)

	_, err = a.routeTriggerEventReplay(ctx, RouteTriggerEventRequest{EventID: eventID, ReplayID: "replay-2"})
	require.NoError(t, err)
	require.NoError(t, db.Where(app.EventDispatch{TriggerEventID: eventID}).Find(&dispatches).Error)
	require.Len(t, dispatches, 2)
}
