package dir

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type sourceFileSetter interface {
	SetSourceFile(path string)
}

type nameFromSourceFileSetter interface {
	SetNameFromSourceFile()
}

func (p *parser) parseDir(path string, typ reflect.Type) (any, error) {
	exists, err := p.fs.DirExists(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check that file exists")
	}
	if !exists {
		return nil, nil
	}

	empty, err := afero.IsEmpty(p.fs, path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check that file is empty")
	}
	if empty {
		return nil, nil
	}

	files, err := p.listDir(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read directory")
	}

	objs := reflect.MakeSlice(typ, 0, len(files))

	for _, f := range files {
		if skipPermissionsPoliciesAsRoles(path, f) {
			continue
		}

		elemType := typ.Elem()
		obj := reflect.New(elemType).Interface()

		parsed, err := p.parseFile(f, path, obj)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse file "+f)
		}

		if !parsed {
			continue
		}

		if !reflect.ValueOf(obj).IsNil() {
			objValue := reflect.ValueOf(obj).Elem()

			if objValue.Kind() == reflect.Ptr && objValue.IsNil() {
				continue
			}

			// why: Set the source file if the object implements sourceFileSetter
			// Note: obj is *T (e.g., *AppPolicy), so we use obj directly for interface checks
			if setter, ok := obj.(sourceFileSetter); ok {
				setter.SetSourceFile(f)
			}

			if setter, ok := obj.(nameFromSourceFileSetter); ok {
				setter.SetNameFromSourceFile()
			}

			objs = reflect.Append(objs, objValue)
		}
	}

	return objs.Interface(), nil
}

func skipPermissionsPoliciesAsRoles(dirPath, filePath string) bool {
	if dirPath != "permissions" {
		return false
	}
	slash := filepath.ToSlash(filePath)
	return slash == "permissions/policies" || strings.HasPrefix(slash, "permissions/policies/")
}
