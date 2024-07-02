package plugins

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

var _ gorm.Plugin = (*viewsPlugin)(nil)

// ViewsPlugin is a plugin that enables turning on a view for specific models. This will overwrite the table name on
// query/preload to add the _view suffix, and use the straight table name for everything else.
func NewViewsPlugin(models []interface{}) *viewsPlugin {
	return &viewsPlugin{
		models:     models,
		viewModels: make(map[string]viewModel, 0),
	}
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
	models     []interface{}
	viewModels map[string]viewModel
}

func (m *viewsPlugin) Name() string {
	return "views-plugin"
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

// see note above
func (m *viewsPlugin) enableView(tx *gorm.DB) {
	schema := tx.Statement.Schema
	vm, ok := m.viewModels[schema.Table]
	if !ok {
		return
	}

	tx.Table(vm.view)
}
