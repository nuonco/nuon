package plugins

import (
	"reflect"

	"gorm.io/gorm"
)

type Tabler interface {
	TableName() string
}

// TableNameOf returns a model's table name from its Tabler implementation, for
// callers without a db handle such as temporal workflow code. Usage:
// TableNameOf[app.Install]().
func TableNameOf[T any, PT interface {
	*T
	Tabler
}]() string {
	return PT(new(T)).TableName()
}

func TableName(db *gorm.DB, obj any) string {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}

	// Check if the object implements Tabler interface
	if tabler, ok := obj.(Tabler); ok {
		return tabler.TableName()
	}

	// Fall back to using the naming strategy
	modelType := reflect.Indirect(value).Type()
	return db.NamingStrategy.TableName(modelType.Name())
}
