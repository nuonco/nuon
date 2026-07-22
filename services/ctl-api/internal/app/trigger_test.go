package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTriggerConstants(t *testing.T) {
	require.Equal(t, TriggerAuthType("hmac"), TriggerAuthTypeHMAC)
	require.Equal(t, EventEnvelopeType("none"), EventEnvelopeTypeNone)
	require.Equal(t, TriggerFilterType("eq"), TriggerFilterTypeEq)
	require.Equal(t, TriggerTargetType("app_branch_run"), TriggerTargetTypeAppBranchRun)
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

func TestTriggersFeatureIsDisabledByDefault(t *testing.T) {
	features := GetFeatures()
	require.Equal(t, OrgFeatureTriggers, features[len(features)-1])

	org := Org{}
	tx := &gorm.DB{Statement: &gorm.Statement{Context: context.Background()}}
	require.NoError(t, org.BeforeCreate(tx))
	require.False(t, org.Features[string(OrgFeatureTriggers)])
}

func TestTriggerFilterPreservesJSONNumberPrecision(t *testing.T) {
	var filters []TriggerFilter
	require.NoError(t, json.Unmarshal([]byte(`[{"op":"eq","path":"/id","value":9007199254740993}]`), &filters))
	require.Len(t, filters, 1)
	require.Equal(t, json.Number("9007199254740993"), filters[0].Value)
}

func TestTriggerFilterAllowsOmittedValue(t *testing.T) {
	var filter TriggerFilter
	require.NoError(t, json.Unmarshal([]byte(`{"op":"exists","path":"$.ref"}`), &filter))
	require.Equal(t, TriggerFilterTypeExists, filter.Op)
	require.Nil(t, filter.Value)
}

func TestTriggerUniqueIndexes(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	sourceIndexes := (&Trigger{}).Indexes(db)
	require.Equal(t, []string{"org_id", "name", "deleted_at"}, sourceIndexes[0].Columns)
	require.True(t, sourceIndexes[0].UniqueValue.Bool)
	require.Equal(t, []string{"ingress_key_hash"}, sourceIndexes[2].Columns)
	require.True(t, sourceIndexes[2].UniqueValue.Bool)

	secretIndexes := (&TriggerSecret{}).Indexes(db)
	require.Equal(t, []string{"trigger_id", "key_id"}, secretIndexes[0].Columns)
	require.True(t, secretIndexes[0].UniqueValue.Bool)

	eventIndexes := (&TriggerEvent{}).Indexes(db)
	require.Equal(t, []string{"trigger_id", "source", "dedupe_id"}, eventIndexes[0].Columns)
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
	require.Equal(t, "WHERE deleted_at = 0 AND name = 'app-triggers'", queueIndexes[2].Option)
}

func TestTriggerSecretDefaultsKeyIDToID(t *testing.T) {
	secret := TriggerSecret{}
	tx := &gorm.DB{Statement: &gorm.Statement{Context: context.Background()}}
	require.NoError(t, secret.BeforeCreate(tx))
	require.NotEmpty(t, secret.ID)
	require.Equal(t, secret.ID, secret.KeyID)
}

func TestTriggerRelationships(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	tests := []struct {
		model         any
		relationships []string
	}{
		{model: &Trigger{}, relationships: []string{"Org", "Secrets", "Events"}},
		{model: &TriggerSecret{}, relationships: []string{"Org", "Trigger"}},
		{model: &TriggerEvent{}, relationships: []string{"Org", "Trigger", "TriggerSecret"}},
		{model: &TriggerRule{}, relationships: []string{"Org", "App", "AppConfig", "Trigger", "AppBranch"}},
		{model: &EventDispatch{}, relationships: []string{"Org", "App", "TriggerEvent", "TriggerRule"}},
		{model: &AppBranchRun{}, relationships: []string{"TriggerEventDispatch"}},
		{model: &InstallRunbookRun{}, relationships: []string{"TriggerEventDispatch"}},
		{model: &EventRunbookWaiter{}, relationships: []string{"MatchedEvent"}},
	}

	for _, test := range tests {
		stmt := &gorm.Statement{DB: db}
		require.NoError(t, stmt.Parse(test.model))
		for _, relationship := range test.relationships {
			require.Contains(t, stmt.Schema.Relationships.Relations, relationship)
		}
	}
}

func TestTriggerCheckConstraints(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	tests := []struct {
		model       any
		constraints []string
	}{
		{model: &Trigger{}, constraints: []string{"trigger_auth_type_checker", "trigger_envelope_checker", "trigger_status_checker"}},
		{model: &TriggerSecret{}, constraints: []string{"trigger_secret_expiration_checker"}},
		{model: &TriggerEvent{}, constraints: []string{"event_routing_status_checker"}},
		{model: &TriggerRule{}, constraints: []string{"trigger_rule_validity_checker", "trigger_rule_target_type_checker", "trigger_rule_target_shape_checker"}},
		{model: &EventDispatch{}, constraints: []string{"event_dispatch_target_type_checker", "event_dispatch_target_shape_checker", "event_dispatch_status_checker", "event_dispatch_attempts_checker"}},
		{model: &EventRunbookWaiter{}, constraints: []string{"event_runbook_waiter_status_checker"}},
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
