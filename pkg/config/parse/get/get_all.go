package get

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	getter "github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/features"
)

func (g *get) GetAll(ctx context.Context) error {
	return g.walkFields(ctx, g.dst, "")
}

func (g *get) walkFields(ctx context.Context, v interface{}, subdir string) error {
	val := reflect.ValueOf(v)

	// If it's a pointer, get the underlying value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

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
		}

		// check if get-enabled exists
		getEnabled, err := features.HasGetFeature(fieldType)
		if err != nil {
			return errors.Wrap(err, "unable to parse field "+fieldType.Name)
		}

		fmt.Println("enabled", getEnabled, fieldType.Name)
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
				newStr := reflect.New(reflect.TypeOf(""))
				newStr.Elem().SetString(val)
				field.Set(newStr)
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

	if err := client.Get(); err != nil {
		return "", errors.Wrap(err, "failed to download file")
	}

	content, err := os.ReadFile(tmpFP)
	if err != nil {
		return "", errors.Wrap(err, "failed to read downloaded file")
	}

	return string(content), nil
}
