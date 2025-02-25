package plugins

import (
	"reflect"

	"gorm.io/gorm"
)

func TableName(db *gorm.DB, obj any) string {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}

	modelType := reflect.Indirect(value).Type()

	return db.NamingStrategy.TableName(modelType.Name())
}
