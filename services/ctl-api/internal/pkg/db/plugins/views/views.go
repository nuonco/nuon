package views

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

var _ gorm.Plugin = (*viewsPlugin)(nil)

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

func (m *viewsPlugin) modelsToViewTables(db *gorm.DB) {
	for _, model := range m.models {
		vm, ok := model.(ViewModel)
		if !ok {
			continue
		}
		if !vm.UseView() {
			continue
		}

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

	tx.Table(vm.view)
}
