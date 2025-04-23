package render

import (
	"reflect"

	"github.com/pkg/errors"
)

func RenderMap(obj any, data map[string]any) error {
	data = EnsurePrefix(data)

	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	iter := val.MapRange()
	for iter.Next() {
		mapValue := iter.Value()

		// If the map value is a string, try to render it
		// Handle different types that can be rendered
		switch mapValue.Kind() {
		case reflect.String:
			strValue := mapValue.String()
			rendered, err := renderStrField(strValue, data)
			if err != nil {
				return errors.Wrap(err, "unable to render string map value")
			}
			val.SetMapIndex(iter.Key(), reflect.ValueOf(rendered))
		case reflect.Map:
			// Recursively handle nested maps
			if err := RenderMap(mapValue.Interface(), data); err != nil {
				return errors.Wrap(err, "unable to render nested map")
			}
		case reflect.Slice:
			// Handle byte slices
			if mapValue.Type().Elem().Kind() == reflect.Uint8 {
				strValue := string(mapValue.Bytes())
				rendered, err := renderStrField(strValue, data)
				if err != nil {
					return errors.Wrap(err, "unable to render []byte map value")
				}
				val.SetMapIndex(iter.Key(), reflect.ValueOf([]byte(rendered)))
			}
		case reflect.Interface:
			// Handle interface{} values
			if !mapValue.IsNil() {
				elem := mapValue.Elem()
				switch elem.Kind() {
				case reflect.Map:
					if err := RenderMap(elem.Interface(), data); err != nil {
						return errors.Wrap(err, "unable to render interface map value")
					}
				case reflect.String:
					strValue := elem.String()
					rendered, err := renderStrField(strValue, data)
					if err != nil {
						return errors.Wrap(err, "unable to render interface string map value")
					}
					val.SetMapIndex(iter.Key(), reflect.ValueOf(rendered))
				case reflect.Slice:
					if elem.Type().Elem().Kind() == reflect.Uint8 {
						strValue := string(elem.Bytes())
						rendered, err := renderStrField(strValue, data)
						if err != nil {
							return errors.Wrap(err, "unable to render interface []byte map value")
						}
						val.SetMapIndex(iter.Key(), reflect.ValueOf([]byte(rendered)))
					}
				}
			}
		}

	}

	return nil
}
