package db

import (
	"reflect"

	"gorm.io/gorm"
)

// AfterQuery is a custom plugin, that allows us to call an after query hook on models, that is not supported by gorm.
// This allows us to do things such as, de-nest data and load it from nested objects into top level pointers, without
// writing a bunch of helper functions and others. It works just like any other gorm hook.
type afterQuery interface {
	AfterQuery(db *gorm.DB) error
}

type afterQueryPlugin struct{}

func (d *afterQueryPlugin) plugin(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	d.callMethod(db, func(value interface{}, tx *gorm.DB) (called bool) {
		if i, ok := value.(afterQuery); ok {
			called = true
			db.AddError(i.AfterQuery(tx))
		}
		return called
	})
}

// This callMethod function, is copied from
// https://raw.githubusercontent.com/go-gorm/gorm/master/callbacks/callmethod.go, which is how the gorm hooks dispatch
// calls to model functions.
func (d *afterQueryPlugin) callMethod(db *gorm.DB, fc func(value interface{}, tx *gorm.DB) bool) {
	tx := db.Session(&gorm.Session{NewDB: true})
	if called := fc(db.Statement.ReflectValue.Interface(), tx); !called {
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			db.Statement.CurDestIndex = 0
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				if value := reflect.Indirect(db.Statement.ReflectValue.Index(i)); value.CanAddr() {
					fc(value.Addr().Interface(), tx)
				} else {
					db.AddError(gorm.ErrInvalidValue)
					return
				}
				db.Statement.CurDestIndex++
			}
		case reflect.Struct:
			if db.Statement.ReflectValue.CanAddr() {
				fc(db.Statement.ReflectValue.Addr().Interface(), tx)
			} else {
				db.AddError(gorm.ErrInvalidValue)
			}
		}
	}
}
