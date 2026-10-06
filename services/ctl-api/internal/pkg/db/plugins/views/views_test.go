package views

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type componentConfigConnection struct{}

func (componentConfigConnection) UseView() bool       { return true }
func (componentConfigConnection) ViewVersion() string { return "v1" }

func testDB() *gorm.DB {
	return &gorm.DB{
		Config: &gorm.Config{
			NamingStrategy: schema.NamingStrategy{},
		},
	}
}

func mappedView(db *gorm.DB) string {
	model := &componentConfigConnection{}
	p := NewViewsPlugin([]interface{}{model})
	p.modelsToViewTables(db)
	return p.viewModels["component_config_connections"].view
}

func TestTableViewOverrideFromEnv(t *testing.T) {
	t.Setenv(EnvOverridePrefix+"component_config_connections", "component_config_connections_view_v2")
	require.Equal(t, "component_config_connections_view_v2", mappedView(testDB()))
}

func TestTableViewOverrideUnset(t *testing.T) {
	require.Equal(t, "component_config_connections_view_v1", mappedView(testDB()))
}

func TestTableViewOverrideBlankEnv(t *testing.T) {
	t.Setenv(EnvOverridePrefix+"component_config_connections", "  ")
	require.Equal(t, "component_config_connections_view_v1", mappedView(testDB()))
}
