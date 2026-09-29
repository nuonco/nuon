package plugins

import (
	"reflect"

	"gorm.io/gorm"
)

type Tabler interface {
	TableName() string
}

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

	if tabler, ok := obj.(Tabler); ok {
		return tabler.TableName()
	}

	modelType := reflect.Indirect(value).Type()
	return db.NamingStrategy.TableName(modelType.Name())
}
