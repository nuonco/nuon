package airgap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
)

func planInputsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE install_inputs (
		id text primary key, created_at datetime, deleted_at integer not null default 0,
		install_id text, "values" text
	)`).Error)
	return db
}

func inputEnvelope(inputs []runnerairgap.InputSpec, plan string) *runnerairgap.Envelope {
	return &runnerairgap.Envelope{
		InstallID: "inl-reference",
		Inputs:    inputs,
		Steps: []runnerairgap.Step{{
			ID:            "step-1",
			CompositePlan: json.RawMessage(plan),
		}},
	}
}

func TestRewriteInputPlaceholdersExactLeaf(t *testing.T) {
	db := planInputsTestDB(t)
	envelope := inputEnvelope([]runnerairgap.InputSpec{{Name: "region", Default: "us-west-2"}}, `{"vars":{"region":"us-west-2"}}`)

	require.NoError(t, rewriteInputPlaceholders(context.Background(), db, envelope))
	require.True(t, envelope.Inputs[0].Bindable)
	require.JSONEq(t, `{"vars":{"region":"__NUON_INPUT_region__"}}`, string(envelope.Steps[0].CompositePlan))
}

func TestRewriteInputPlaceholdersEmbeddedValue(t *testing.T) {
	db := planInputsTestDB(t)
	envelope := inputEnvelope([]runnerairgap.InputSpec{{Name: "domain", Default: "example.com"}}, `{"url":"https://api.example.com/v1"}`)

	require.NoError(t, rewriteInputPlaceholders(context.Background(), db, envelope))
	require.True(t, envelope.Inputs[0].Bindable)
	require.JSONEq(t, `{"url":"https://api.__NUON_INPUT_domain__/v1"}`, string(envelope.Steps[0].CompositePlan))
}

func TestRewriteInputPlaceholdersSkipsGenericValues(t *testing.T) {
	db := planInputsTestDB(t)
	envelope := inputEnvelope([]runnerairgap.InputSpec{
		{Name: "replicas", Default: "3"},
		{Name: "enabled", Default: "true"},
	}, `{"replicas":"3","enabled":"true"}`)

	require.NoError(t, rewriteInputPlaceholders(context.Background(), db, envelope))
	require.False(t, envelope.Inputs[0].Bindable)
	require.False(t, envelope.Inputs[1].Bindable)
	require.JSONEq(t, `{"replicas":"3","enabled":"true"}`, string(envelope.Steps[0].CompositePlan))
}

func TestRewriteInputPlaceholdersRejectsBakedSecret(t *testing.T) {
	db := planInputsTestDB(t)
	envelope := inputEnvelope([]runnerairgap.InputSpec{{Name: "token", Default: "secret-value", Secret: true}}, `{"authorization":"Bearer secret-value"}`)

	err := rewriteInputPlaceholders(context.Background(), db, envelope)
	require.ErrorContains(t, err, `input "token" is a secret`)
}

func TestRewriteInputPlaceholdersUsesLatestInstallValue(t *testing.T) {
	db := planInputsTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO install_inputs (id, created_at, install_id, "values") VALUES (?, ?, ?, ?)`,
		"inputs-old", time.Now().Add(-time.Hour), "inl-reference", `"region"=>"old-region"`).Error)
	require.NoError(t, db.Exec(`INSERT INTO install_inputs (id, created_at, install_id, "values") VALUES (?, ?, ?, ?)`,
		"inputs-new", time.Now(), "inl-reference", `"region"=>"eu-central-1"`).Error)
	envelope := inputEnvelope([]runnerairgap.InputSpec{{Name: "region", Default: "us-west-2"}}, `{"region":"eu-central-1"}`)

	require.NoError(t, rewriteInputPlaceholders(context.Background(), db, envelope))
	require.True(t, envelope.Inputs[0].Bindable)
	require.JSONEq(t, `{"region":"__NUON_INPUT_region__"}`, string(envelope.Steps[0].CompositePlan))
}

func TestRewriteInputPlaceholdersRejectsDuplicateValues(t *testing.T) {
	db := planInputsTestDB(t)
	envelope := inputEnvelope([]runnerairgap.InputSpec{
		{Name: "first", Default: "duplicate-value"},
		{Name: "second", Default: "duplicate-value"},
	}, `{"value":"duplicate-value"}`)

	err := rewriteInputPlaceholders(context.Background(), db, envelope)
	require.ErrorContains(t, err, `inputs "first" and "second" have the same reference value`)
}

func TestRewriteInputPlaceholdersNoInputConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	envelope := inputEnvelope(nil, `{"value":"unchanged"}`)

	require.NoError(t, rewriteInputPlaceholders(context.Background(), db, envelope))
	require.JSONEq(t, `{"value":"unchanged"}`, string(envelope.Steps[0].CompositePlan))
}

func TestExportInputSpecsIncludesDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	for _, ddl := range []string{
		`CREATE TABLE app_input_configs (id text primary key, app_config_id text, deleted_at integer not null default 0)`,
		`CREATE TABLE app_inputs (
			id text primary key, app_input_config_id text, created_at datetime, deleted_at integer not null default 0,
			name text, type text, description text, sensitive integer, required integer, "default" text, "index" integer
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	require.NoError(t, db.Exec(`INSERT INTO app_input_configs (id, app_config_id) VALUES ('aic-1', 'apc-1')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO app_inputs (id, app_input_config_id, name, type, "default", "index") VALUES ('ain-1', 'aic-1', 'region', 'string', 'us-west-2', 0)`).Error)

	specs, err := exportInputSpecs(context.Background(), db, "apc-1")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	require.Equal(t, "us-west-2", specs[0].Default)
}
