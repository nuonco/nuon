package views

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type overrideModel struct{}

func (overrideModel) TableName() string   { return "component_config_connections" }
func (overrideModel) UseView() bool       { return true }
func (overrideModel) ViewVersion() string { return "v1" }

func testDB() *gorm.DB {
	return &gorm.DB{
		Config: &gorm.Config{
			NamingStrategy: schema.NamingStrategy{},
			Plugins:        map[string]gorm.Plugin{},
		},
	}
}

func attachPlugin(db *gorm.DB, p *viewsPlugin) {
	p.modelsToViewTables(db)
	db.Config.Plugins[PluginName] = p
}

func TestNameOverrideDefaultViewFromEnv(t *testing.T) {
	t.Setenv(EnvOverridePrefix+"component_config_connections_view_v1", "component_config_connections_view_v2")

	db := testDB()
	model := &overrideModel{}
	attachPlugin(db, NewViewsPlugin([]interface{}{model}))

	require.Equal(t, "component_config_connections_view_v2", TableOrViewName(db, model, ""))
	require.Equal(t, "component_config_connections_view_v2", CurrentViewName(db, model))
	require.Equal(t, "component_config_connections_view_v2", RewriteName(db, "component_config_connections_view_v1"))
}

func TestNameOverrideLatestView(t *testing.T) {
	db := testDB()
	model := &overrideModel{}
	p := NewViewsPlugin([]interface{}{model}, WithNameOverrides(map[string]string{
		"component_config_connections_latest_configs_view": "component_config_connections_latest_configs_view_v2",
	}))
	attachPlugin(db, p)

	require.Equal(t, "component_config_connections_view_v1", TableOrViewName(db, model, ""))
	require.Equal(t, "component_config_connections_latest_configs_view_v2",
		RewriteName(db, "component_config_connections_latest_configs_view"))
}

func TestMergeNameOverrides(t *testing.T) {
	got := mergeNameOverrides(nil, map[string]string{
		" Component_Config_Connections_View_V1 ": "component_config_connections_view_v2",
		"":                                       "x",
		"latest":                                 "  ",
	})
	require.Equal(t, map[string]string{
		"component_config_connections_view_v1": "component_config_connections_view_v2",
	}, got)
}
