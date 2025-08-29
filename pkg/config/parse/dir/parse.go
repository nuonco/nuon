package dir

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type fieldOpts struct {
	Name     string
	Nonempty bool
	Required bool
}

func (p *parser) parseField(str string) (*fieldOpts, error) {
	pieces := strings.Split(str, ",")
	if len(pieces) < 1 {
		return nil, errors.New("invalid field tag")
	}
	nonempty := generics.SliceContains("nonempty", pieces[1:])
	required := generics.SliceContains("required", pieces[1:])

	return &fieldOpts{
		Name:     pieces[0],
		Nonempty: nonempty,
		Required: required,
	}, nil
}

func (p *parser) parse(ctx context.Context) error {
	v := reflect.ValueOf(p.dst)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a struct, got %T", p.dst)
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Get the filename from the name tag
		arg, ok := field.Tag.Lookup("name")
		if !ok {
			continue
		}

		fieldOpts, err := p.parseField(arg)
		if err != nil {
			return err
		}

		if field.Type.Kind() != reflect.Slice {
			// For regular fields, load the file directly
			// Create a pointer to the field's type
			// If field.Type is already a pointer, get its element type
			elemType := field.Type
			if field.Type.Kind() == reflect.Ptr {
				elemType = field.Type.Elem()
			}

			obj := reflect.New(elemType).Interface()
			parsed, err := p.parseFile(fieldOpts.Name, obj)
			if err != nil {
				return errors.Wrap(err, "unable to load file for "+fieldOpts.Name)
			}

			if !parsed {
				if fieldOpts.Required {
					return fmt.Errorf("missing required file %s", fieldOpts.Name+p.opts.Ext)
				}
				continue
			}

			if fieldValue.CanSet() && !reflect.ValueOf(obj).IsZero() {
				fieldValue.Set(reflect.ValueOf(obj))
			}
		}
		if field.Type.Kind() == reflect.Slice {
			objs, err := p.parseDir(fieldOpts.Name, field.Type)
			if err != nil {
				return errors.Wrap(err, "unable to load subdir "+fieldOpts.Name)
			}

			if objs != nil && fieldValue.CanSet() {
				// obj is already a pointer, so we can set it directly
				fieldValue.Set(reflect.ValueOf(objs))
			}
		}
	}

	return nil
}
