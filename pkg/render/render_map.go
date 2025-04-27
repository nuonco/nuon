package render

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

func RenderMap(obj any, data map[string]any) error {
	data = EnsurePrefix(data)

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer to a map")
	}

	val = val.Elem()
	// If we have an interface, we need to get the concrete value
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() != reflect.Map {
		return errors.New("underlying type must be a map")
	}

	iter := val.MapRange()
	for iter.Next() {
		mapValue := iter.Value()

		// If the map value is a string, try to render it
		// Handle different types that can be rendered
		switch mapValue.Kind() {
		case reflect.String:
			strValue := mapValue.String()
			rendered, err := RenderV2(strValue, data)
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
			if mapValue.IsNil() {
				continue
			}
			elem := mapValue.Elem()
			// If we have a pointer, dereference it first
			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}
			switch elem.Kind() {
			case reflect.Map:
				// Create a pointer to the map value
				mapPtr := reflect.New(elem.Type())
				mapPtr.Elem().Set(elem)
				if err := RenderMap(mapPtr.Interface(), data); err != nil {
					return errors.Wrap(err, "unable to render interface map value")
				}

				// Update the original map with the rendered value
				val.SetMapIndex(iter.Key(), mapPtr.Elem())
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
			default:
				return fmt.Errorf("unsupported type %T %s", elem, elem.Kind())
			}
		}

	}

	return nil
}
