package render

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/render/features"
)

// want to write a type that can walk an object recursively and any field that has a struct
func RenderStruct(obj any, data map[string]any) error {
	return walkFields(obj, data)
}

func walkFields(obj any, data map[string]any) error {
	val := reflect.ValueOf(obj)

	// If it's a pointer, get the underlying value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Map {
		return RenderMap(obj, data)
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		enabled, err := features.HasTemplateFeature(fieldType)
		if err != nil {
			return errors.Wrap(err, "unable to check if feature is enabled")
		}

		// if the record is nested, recurse
		switch field.Kind() {
		case reflect.Ptr:
			// If it's a nil pointer, skip it
			if field.IsNil() {
				continue
			}
			if err := walkFields(field.Interface(), data); err != nil {
				return err
			}
		case reflect.Struct:
			if err := walkFields(field.Addr().Interface(), data); err != nil {
				return err
			}
		case reflect.Map:
			// For maps, iterate through all values
			if !field.CanSet() {
				return errors.New("map field is not settable")
			}
			if !enabled {
				continue
			}

			if err := RenderMap(field.Interface(), data); err != nil {
				return errors.Wrap(err, "unable to render map")
			}
		case reflect.Slice:
			// Handle slices of structs
			elemKind := field.Type().Elem().Kind()

			if elemKind == reflect.Struct {
				for i := 0; i < field.Len(); i++ {
					elem := field.Index(i)
					if err := walkFields(elem.Addr().Interface(), data); err != nil {
						return err
					}
				}
			} else if elemKind == reflect.Ptr && field.Type().Elem().Elem().Kind() == reflect.Struct {
				// Handle slice of pointers to structs
				for i := 0; i < field.Len(); i++ {
					elem := field.Index(i)
					if elem.IsNil() {
						continue
					}
					if err := walkFields(elem.Interface(), data); err != nil {
						return err
					}
				}
			} else if elemKind == reflect.Uint8 {
				byteValue := field.Bytes()

				val, err := renderByteField(byteValue, data)
				if err != nil {
					return errors.Wrap(err, "unable to fetch field value")
				}

				if !field.CanSet() {
					return errors.New("field is not settable: " + fieldType.Name)
				}

				field.SetBytes(val)
			}
		case reflect.String:
			if !enabled {
				continue
			}

			val, err := renderStrField(field.String(), data)
			if err != nil {
				return errors.Wrap(err, "unable to fetch field value")
			}

			if !field.CanSet() {
				return errors.New("field is not settable: " + fieldType.Name)
			}

			if field.Kind() == reflect.Ptr {
				newStr := reflect.New(reflect.TypeOf(""))
				newStr.Elem().SetString(val)
				field.Set(newStr)
			} else {
				field.SetString(val)
			}
		default:
			if !enabled {
				continue
			}

			return errors.New("invalid type to render features on")
		}
	}

	return nil
}

func renderStrField(inputVal string, data map[string]any) (string, error) {
	data = EnsurePrefix(data)

	return RenderV2(inputVal, data)
}

func renderByteField(inputVal []byte, data map[string]any) ([]byte, error) {
	data = EnsurePrefix(data)

	final, err := RenderV2(string(inputVal), data)
	if err != nil {
		return nil, err
	}

	return []byte(final), nil
}
