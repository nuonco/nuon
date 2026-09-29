package render

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/render/features"
)

func RenderStruct(obj any, data map[string]any) error {
	return walkFields(obj, data)
}

func walkFields(obj any, data map[string]any) error {
	val := reflect.ValueOf(obj)

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

		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}
			elem := field.Elem()
			switch elem.Kind() {
			case reflect.Struct, reflect.Map:
				if err := walkFields(field.Interface(), data); err != nil {
					return err
				}
			case reflect.String:
				if !enabled {
					continue
				}
				val, err := renderStrField(elem.String(), data)
				if err != nil {
					return errors.Wrap(err, "unable to render pointer string field")
				}
				elem.SetString(val)
			default:
				continue
			}
		case reflect.Struct:
			if err := walkFields(field.Addr().Interface(), data); err != nil {
				return err
			}
		case reflect.Map:
			if !field.CanSet() {
				return errors.New("map field is not settable")
			}
			if !enabled {
				continue
			}

			if field.Kind() == reflect.Map {
				if err := RenderMap(field.Addr().Interface(), data); err != nil {
					return errors.Wrap(err, "unable to render map")
				}
			} else {
				if err := RenderMap(field.Interface(), data); err != nil {
					return errors.Wrap(err, "unable to render map")
				}
			}
		case reflect.Slice:
			elemKind := field.Type().Elem().Kind()

			if elemKind == reflect.Struct {
				for i := 0; i < field.Len(); i++ {
					elem := field.Index(i)
					if err := walkFields(elem.Addr().Interface(), data); err != nil {
						return err
					}
				}
			} else if elemKind == reflect.Ptr && field.Type().Elem().Elem().Kind() == reflect.Struct {
				for i := 0; i < field.Len(); i++ {
					elem := field.Index(i)
					if elem.IsNil() {
						continue
					}
					if err := walkFields(elem.Interface(), data); err != nil {
						return err
					}
				}
			} else if elemKind == reflect.String {
				if !enabled {
					continue
				}

				for i := 0; i < field.Len(); i++ {
					elem := field.Index(i)
					val, err := renderStrField(elem.String(), data)
					if err != nil {
						return errors.Wrap(err, "unable to render string in slice")
					}

					if !elem.CanSet() {
						return errors.New("string element in slice is not settable")
					}

					elem.SetString(val)
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

// why: Config fields walked by RenderStruct/RenderMap end up in infrastructure APIs --
// helm values files, kubernetes manifests, terraform variables, env vars, nested
// stack parameters -- never in a browser. They therefore render through
// RenderTextV2: html/template escaping silently corrupts values, e.g. a PEM
// public key inlined into a helm values file loses every "+" to "&#43;".
func renderStrField(inputVal string, data map[string]any) (string, error) {
	data = EnsurePrefix(data)

	return RenderTextV2(inputVal, data)
}

func renderByteField(inputVal []byte, data map[string]any) ([]byte, error) {
	data = EnsurePrefix(data)

	final, err := RenderTextV2(string(inputVal), data)
	if err != nil {
		return nil, err
	}

	return []byte(final), nil
}
