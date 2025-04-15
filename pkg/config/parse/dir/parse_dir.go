package dir

import (
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

func (p *parser) parseDir(path string, typ reflect.Type, required, nonempty bool) (any, error) {
	exists, err := p.fs.DirExists(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check that file exists")
	}
	if !exists {
		if required {
			return nil, ErrMissingDir{
				Name: path,
			}
		}

		return nil, nil
	}

	empty, err := afero.IsEmpty(p.fs, path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check that file is empty")
	}
	if empty && nonempty {
		return nil, ErrMissingFile{
			Name: path,
		}
	}

	files, err := p.listDir(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read directory")
	}

	objs := reflect.MakeSlice(typ, 0, len(files))

	for _, f := range files {
		elemType := typ.Elem()
		obj := reflect.New(elemType).Interface()

		if err := p.parseFile(filepath.Join(path, f), obj, required, nonempty); err != nil {
			return nil, err
		}

		objValue := reflect.ValueOf(obj).Elem()
		objs = reflect.Append(objs, objValue)
	}

	return objs.Interface(), nil
}
