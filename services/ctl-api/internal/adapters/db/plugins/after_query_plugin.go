package plugins

import (
	"gorm.io/gorm"
)

// AfterQuery is a custom plugin, that allows us to call an after query hook on models, that is not supported by gorm.
// This allows us to do things such as, de-nest data and load it from nested objects into top level pointers, without
// writing a bunch of helper functions and others. It works just like any other gorm hook.
type afterQuery interface {
	AfterQuery(db *gorm.DB) error
}

var _ gorm.Plugin = (*afterQueryPlugin)(nil)

func NewAfterQueryPlugin() *afterQueryPlugin {
	return &afterQueryPlugin{}
}

type afterQueryPlugin struct{}

func (d *afterQueryPlugin) Name() string {
	return "after-query"
}

func (d *afterQueryPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Query().After("gorm:query").Register("after_query", d.plugin)

	return nil
}

func (d *afterQueryPlugin) plugin(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	callObjMethod(db, func(value interface{}, tx *gorm.DB) (called bool) {
		if i, ok := value.(afterQuery); ok {
			called = true
			db.AddError(i.AfterQuery(tx))
		}
		return called
	})
}
