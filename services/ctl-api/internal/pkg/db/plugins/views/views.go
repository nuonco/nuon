package views

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

const PluginName = "views-plugin"

// EnvOverridePrefix pins a queried relation to a different relation by full
// name, without changing GORM models. The suffix is the source view name,
// lowercase. Example:
//
//	GORM_VIEW_PLUGIN_OVERRIDE_component_config_connections_view_v1=component_config_connections_view_v2
//	GORM_VIEW_PLUGIN_OVERRIDE_component_config_connections_latest_configs_view=component_config_connections_latest_configs_view_v2
const EnvOverridePrefix = "GORM_VIEW_PLUGIN_OVERRIDE_"

var _ gorm.Plugin = (*viewsPlugin)(nil)

type PluginOption func(*viewsPlugin)

// WithNameOverrides replaces queried relation names. Intended for tests;
// production uses env vars.
func WithNameOverrides(overrides map[string]string) PluginOption {
	return func(p *viewsPlugin) {
		p.nameOverrides = mergeNameOverrides(p.nameOverrides, overrides)
	}
}

// ViewsPlugin is a plugin that enables turning on a view for specific models. This will overwrite the table name on
// query/preload to add the _view suffix, and use the straight table name for everything else.
func NewViewsPlugin(models []interface{}, opts ...PluginOption) *viewsPlugin {
	p := &viewsPlugin{
		models:     models,
		viewModels: make(map[string]viewModel, 0),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type ViewModel interface {
	UseView() bool
	ViewVersion() string
}

type viewModel struct {
	model interface{}
	table string
	view  string
}

type viewsPlugin struct {
	models        []interface{}
	viewModels    map[string]viewModel
	nameOverrides map[string]string
}

func (m *viewsPlugin) Name() string {
	return PluginName
}

func (m *viewsPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().Before("gorm:query").Register("enable_view_on_query", m.enableView)
	db.Callback().Query().Before("gorm:preload").Register("enable_view_on_preload", m.enableView)

	m.modelsToViewTables(db)

	return nil
}

// modelsToViewModels walks through each model, and checks to see if the `UseView` function is set and returns true. It
// builds a map of all view models by table name
func (m *viewsPlugin) modelsToViewTables(db *gorm.DB) {
	for _, model := range m.models {
		vm, ok := model.(ViewModel)
		if !ok {
			continue
		}
		if !vm.UseView() {
			continue
		}

		// this block accepts an interface that points to a model, and turns it into a table name. We probably
		// don't need to be this robust, but it prevents us from passing invalid types in here and having silent
		// errors.
		value := reflect.ValueOf(model)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			value = reflect.New(value.Type().Elem())
		}
		modelType := reflect.Indirect(value).Type()
		if modelType.Kind() == reflect.Interface {
			modelType = reflect.Indirect(reflect.ValueOf(model)).Elem().Type()
		}
		for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		tableName := db.NamingStrategy.TableName(modelType.Name())
		m.viewModels[tableName] = viewModel{
			model: model,
			table: tableName,
			view:  fmt.Sprintf("%s_view_%s", tableName, vm.ViewVersion()),
		}
	}
}

func (m *viewsPlugin) rewriteName(name string) string {
	if name == "" {
		return name
	}
	key := strings.ToLower(name)
	if m != nil {
		if to, ok := m.nameOverrides[key]; ok && to != "" {
			return to
		}
	}
	if to := strings.TrimSpace(os.Getenv(EnvOverridePrefix + key)); to != "" {
		return to
	}
	return name
}

func pluginFromDB(db *gorm.DB) *viewsPlugin {
	if db == nil || db.Config == nil {
		return nil
	}
	p, ok := db.Config.Plugins[PluginName].(*viewsPlugin)
	if !ok {
		return nil
	}
	return p
}

func mergeNameOverrides(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for from, to := range base {
		out[from] = to
	}
	for from, to := range extra {
		from = strings.ToLower(strings.TrimSpace(from))
		to = strings.TrimSpace(to)
		if from == "" || to == "" {
			continue
		}
		out[from] = to
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// see note above
func (m *viewsPlugin) enableView(tx *gorm.DB) {
	disable, ok := tx.InstanceGet(DisableViewsKey)
	if ok && disable.(bool) {
		return
	}

	schema := tx.Statement.Schema
	if schema == nil {
		return
	}
	vm, ok := m.viewModels[schema.Table]
	if !ok {
		return
	}

	tx.Table(m.rewriteName(vm.view))
}
