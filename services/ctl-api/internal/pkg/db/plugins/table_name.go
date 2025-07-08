package plugins

import (
	"reflect"

	"gorm.io/gorm"
)

type Tabler interface {
	TableName() string
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
