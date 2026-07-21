package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestEventAutomationConstants(t *testing.T) {
	require.Equal(t, EventSourceType("generic_hmac"), EventSourceTypeGenericHMAC)
	require.Equal(t, EventAutomationFilterType("eq"), EventAutomationFilterTypeEq)
	require.Equal(t, EventAutomationTargetType("app_branch_run"), EventAutomationTargetTypeAppBranchRun)
	require.ElementsMatch(t, []EventDispatchStatus{
		"pending", "dispatching", "triggered", "retryable_failed", "dead_lettered", "cancelled",
	}, []EventDispatchStatus{
		EventDispatchStatusPending,
		EventDispatchStatusDispatching,
		EventDispatchStatusTriggered,
		EventDispatchStatusRetryableFailed,
		EventDispatchStatusDeadLettered,
		EventDispatchStatusCancelled,
	})
}

func TestEventAutomationUniqueIndexes(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	sourceIndexes := (&EventSource{}).Indexes(db)
	require.Equal(t, []string{"app_id", "name", "deleted_at"}, sourceIndexes[0].Columns)
	require.True(t, sourceIndexes[0].UniqueValue.Bool)
	require.Equal(t, []string{"ingress_key_hash"}, sourceIndexes[3].Columns)
	require.True(t, sourceIndexes[3].UniqueValue.Bool)

	secretIndexes := (&EventSourceSecret{}).Indexes(db)
	require.Equal(t, []string{"event_source_id", "key_id"}, secretIndexes[0].Columns)
	require.True(t, secretIndexes[0].UniqueValue.Bool)

	eventIndexes := (&EventSourceEvent{}).Indexes(db)
	require.Equal(t, []string{"event_source_id", "external_id"}, eventIndexes[0].Columns)
	require.True(t, eventIndexes[0].UniqueValue.Bool)

	dispatchIndexes := (&EventDispatch{}).Indexes(db)
	require.Equal(t, []string{"idempotency_key"}, dispatchIndexes[0].Columns)
	require.True(t, dispatchIndexes[0].UniqueValue.Bool)

	queueSignalIndexes := (&QueueSignal{}).Indexes(db)
	require.Equal(t, []string{"queue_id", "dedupe_key"}, queueSignalIndexes[0].Columns)
	require.True(t, queueSignalIndexes[0].UniqueValue.Bool)
	require.Equal(t, "WHERE deleted_at = 0 AND dedupe_key IS NOT NULL AND dedupe_key <> ''", queueSignalIndexes[0].Option)

	queueIndexes := (&Queue{}).Indexes(db)
	require.Equal(t, []string{"owner_id", "owner_type", "name"}, queueIndexes[2].Columns)
	require.True(t, queueIndexes[2].UniqueValue.Bool)
	require.Equal(t, "WHERE deleted_at = 0 AND name = 'app-automations'", queueIndexes[2].Option)
}

func TestEventSourceSecretDefaultsKeyIDToID(t *testing.T) {
	secret := EventSourceSecret{}
	tx := &gorm.DB{Statement: &gorm.Statement{Context: context.Background()}}
	require.NoError(t, secret.BeforeCreate(tx))
	require.NotEmpty(t, secret.ID)
	require.Equal(t, secret.ID, secret.KeyID)
}

func TestEventAutomationRelationships(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	tests := []struct {
		model         any
		relationships []string
	}{
		{model: &EventSource{}, relationships: []string{"Org", "App", "Secrets", "Events"}},
		{model: &EventSourceSecret{}, relationships: []string{"Org", "EventSource"}},
		{model: &EventSourceEvent{}, relationships: []string{"Org", "App", "EventSource", "EventSourceSecret"}},
		{model: &EventAutomationRule{}, relationships: []string{"Org", "App", "AppConfig", "EventSource", "AppBranch"}},
		{model: &EventDispatch{}, relationships: []string{"Org", "App", "EventSourceEvent", "EventAutomationRule"}},
		{model: &AppBranchRun{}, relationships: []string{"AutomationDispatch"}},
	}

	for _, test := range tests {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(test.model))
		for _, relationship := range test.relationships {
			require.Contains(t, stmt.Schema.Relationships.Relations, relationship)
		}
	}
}

func TestEventAutomationCheckConstraints(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	tests := []struct {
		model       any
		constraints []string
	}{
		{model: &EventSource{}, constraints: []string{"event_source_type_checker", "event_source_status_checker"}},
		{model: &EventSourceSecret{}, constraints: []string{"event_source_secret_expiration_checker"}},
		{model: &EventSourceEvent{}, constraints: []string{"event_routing_status_checker"}},
		{model: &EventAutomationRule{}, constraints: []string{"event_automation_rule_validity_checker", "event_automation_rule_event_types_checker", "event_automation_rule_target_type_checker"}},
		{model: &EventDispatch{}, constraints: []string{"event_dispatch_target_type_checker", "event_dispatch_status_checker", "event_dispatch_attempts_checker"}},
	}

	for _, test := range tests {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(test.model))
		constraints := stmt.Schema.ParseCheckConstraints()
		for _, constraint := range test.constraints {
			require.Contains(t, constraints, constraint)
		}
	}
}
