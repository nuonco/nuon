package app

import (
	"reflect"
	"sync"
	"testing"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestActionWorkflowTriggerConfigSize(t *testing.T) {
	size := reflect.TypeOf(ActionWorkflowTriggerConfig{}).Size()
	t.Logf("trigger size: %d bytes; 120 ten-slot slices: %.2f MiB", size, float64(120*10*size)/(1024*1024))
	// GORM reserves ten trigger values per action even when only one trigger is loaded.
	require.Less(t, size, uintptr(8*1024))
}

func TestActionWorkflowTriggerConfigRelationships(t *testing.T) {
	s, err := schema.Parse(&ActionWorkflowTriggerConfig{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	for _, name := range []string{"App", "AppConfig", "ActionWorkflowConfig"} {
		t.Run(name, func(t *testing.T) {
			rel := s.Relationships.Relations[name]
			require.NotNil(t, rel)
			require.Equal(t, schema.BelongsTo, rel.Type)
			require.Len(t, rel.References, 1)
			require.Equal(t, name+"ID", rel.References[0].ForeignKey.Name)
			require.Equal(t, "ID", rel.References[0].PrimaryKey.Name)
		})
	}
	parent := s.Relationships.Relations["ActionWorkflowConfig"].FieldSchema
	rel := parent.Relationships.Relations["Triggers"]
	require.NotNil(t, rel)
	require.Equal(t, schema.HasMany, rel.Type)
	require.Equal(t, "ActionWorkflowConfigID", rel.References[0].ForeignKey.Name)
}

func TestActionWorkflowTriggerConfigNestedCreate(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	var tables []string
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("capture_tables", func(tx *gorm.DB) {
		tables = append(tables, tx.Statement.Table)
	}))
	cfg := ActionWorkflowConfig{
		AppID: "app-parent", AppConfigID: "config-parent", ActionWorkflowID: "workflow-parent",
		Triggers: []ActionWorkflowTriggerConfig{{
			AppID: "app-parent", AppConfigID: "config-parent", Type: ActionWorkflowTriggerTypeManual,
		}},
	}
	require.NoError(t, db.Create(&cfg).Error)
	require.ElementsMatch(t, []string{"action_workflow_configs", "action_workflow_trigger_configs"}, tables)
	require.NotEmpty(t, cfg.ID)
	require.Equal(t, cfg.ID, cfg.Triggers[0].ActionWorkflowConfigID)
	require.Equal(t, "app-parent", cfg.Triggers[0].AppID)
	require.Equal(t, "config-parent", cfg.Triggers[0].AppConfigID)
	require.Nil(t, cfg.Triggers[0].App)
	require.Nil(t, cfg.Triggers[0].AppConfig)
	require.Nil(t, cfg.Triggers[0].ActionWorkflowConfig)
}

func TestActionWorkflowTriggerConfigSelection(t *testing.T) {
	cfg := ActionWorkflowConfig{Triggers: []ActionWorkflowTriggerConfig{
		{ID: "manual", Type: ActionWorkflowTriggerTypeManual, AppID: "app", AppConfigID: "config", ActionWorkflowConfigID: "action"},
		{ID: "cron", Type: ActionWorkflowTriggerTypeCron, CronSchedule: "17 3 * * *"},
		{ID: "component", Type: ActionWorkflowTriggerTypePostDeployComponent, ComponentID: generics.NewNullString("component-a"), Index: 7},
	}}
	require.NoError(t, cfg.AfterQuery(nil))
	got := &cfg
	require.Len(t, got.Triggers, 3)
	require.Equal(t, "app", got.Triggers[0].AppID)
	require.Equal(t, "config", got.Triggers[0].AppConfigID)
	require.Equal(t, "action", got.Triggers[0].ActionWorkflowConfigID)
	require.True(t, got.WorkflowConfigCanTriggerManually())
	require.NotNil(t, got.CronTrigger)
	require.Equal(t, "17 3 * * *", got.CronTrigger.CronSchedule)
	require.Len(t, got.LifecycleTriggers, 1)
	require.Equal(t, "component", got.LifecycleTriggers[0].ID)
	require.True(t, got.HasComponentTrigger(ActionWorkflowTriggerTypePostDeployComponent, "component-a"))
	require.False(t, got.HasComponentTrigger(ActionWorkflowTriggerTypePostDeployComponent, "component-b"))
	require.Equal(t, 7, got.GetComponentTriggerIndex(ActionWorkflowTriggerTypePostDeployComponent, "component-a"))
}
