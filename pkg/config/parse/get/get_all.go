package get

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/features"
)

func (g *get) GetAll(ctx context.Context) error {
	return g.walkFields(ctx, g.dst, "")
}

func (g *get) walkFields(ctx context.Context, v interface{}, subdir string) error {
	val := reflect.ValueOf(v)

	// If it's not a pointer, we need to get a pointer to make it settable
	if val.Kind() != reflect.Ptr {
		// Create a new pointer to the value
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}

	// Now we can safely get the underlying value
	val = val.Elem()

	// We only process struct types
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// if the record is nested, recurse
		switch field.Kind() {
		case reflect.Ptr:
			// If it's a nil pointer, skip it
			if field.IsNil() {
				continue
			}

			if fieldType.Name == "Components" {
				subdir = "components"
			}

			if fieldType.Name == "Actions" {
				subdir = "actions"
			}

			// Recurse with the dereferenced pointer
			if err := g.walkFields(ctx, field.Interface(), subdir); err != nil {
				return err
			}
		case reflect.Struct:
			if fieldType.Name == "Components" {
				subdir = "components"
			}

			if fieldType.Name == "Actions" {
				subdir = "actions"
			}

			if err := g.walkFields(ctx, field.Interface(), subdir); err != nil {
				return err
			}
		case reflect.Slice:
			if fieldType.Name == "Components" {
				subdir = "components"
			}

			if fieldType.Name == "Actions" {
				subdir = "actions"
			}

			// Handle slices of structs
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				switch elem.Kind() {
				case reflect.Struct:
					// Create a pointer to make it settable
					ptr := reflect.New(elem.Type())
					ptr.Elem().Set(elem)
					if err := g.walkFields(ctx, ptr.Interface(), subdir); err != nil {
						return err
					}
					// Update the slice element with potentially modified value
					if elem.CanSet() {
						elem.Set(ptr.Elem())
					}
				case reflect.Ptr:
					if elem.IsNil() {
						continue
					}

					if elem.Elem().Kind() == reflect.Struct {
						if err := g.walkFields(ctx, elem.Interface(), subdir); err != nil {
							return err
						}
					}
				}
			}
		}

		// check if get-enabled exists
		getEnabled, err := features.HasGetFeature(fieldType)
		if err != nil {
			return errors.Wrap(err, "unable to parse field "+fieldType.Name)
		}

		if !getEnabled {
			continue
		}

		// now, check if it is a string field
		if field.Kind() == reflect.String {
			strValue := field.String()

			val, err := g.processField(ctx, strValue, subdir)
			if err != nil {
				return errors.Wrap(err, "unable to fetch field value")
			}

			if !field.CanSet() {
				return errors.New("field is not settable: " + fieldType.Name)
			}

			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					// Create a new pointer if it's nil
					field.Set(reflect.New(field.Type().Elem()))
				}
				// Set the value on the element that the pointer points to
				field.Elem().SetString(val)
			} else {
				field.SetString(val)
			}

		} else {
			return errors.New("get feature enabled on a non-string field " + fieldType.Name)
		}
	}

	return nil
}

func (g *get) processField(ctx context.Context, inputVal string, subdir string) (string, error) {
	prefixes := []string{
		"http",
		"./",
		"git",
	}
	isGettable := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(inputVal, prefix) {
			isGettable = true
		}
	}

	if strings.HasPrefix(inputVal, "./nuon") {
		return inputVal, nil
	}

	if !isGettable {
		return inputVal, nil
	}

	pwd := filepath.Join(g.opts.RootDir, subdir)

	if _, err := getter.Detect(inputVal, pwd, GetDetectors()); err != nil {
		return inputVal, nil
	}

	// Create a temporary directory to store the downloaded file
	tmpDir, err := os.MkdirTemp(pwd, ".nuon-get-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	tmpFP := filepath.Join(tmpDir, "field")

	ctx, cancel := context.WithTimeout(ctx, g.opts.FieldTimeout)
	defer cancel()

	// Configure the client
	client := &getter.Client{
		Ctx:  ctx,
		Src:  inputVal,
		Dir:  true,
		Dst:  tmpFP,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
	}

	fmt.Println(tmpFP, pwd, inputVal)
	if err := client.Get(); err != nil {
		return "", errors.Wrap(err, "failed to download file")
	}

	content, err := os.ReadFile(tmpFP)
	if err != nil {
		return "", errors.Wrap(err, "failed to read downloaded file")
	}

	return string(content), nil
}
