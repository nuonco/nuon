package dir

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

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
		elemType := typ.Elem()
		obj := reflect.New(elemType).Interface()

		parsed, err := p.parseFile(f, obj)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse file "+f)
		}

		if !parsed {
			continue
		}

		// Only append non-nil objects
		if !reflect.ValueOf(obj).IsNil() {
			objValue := reflect.ValueOf(obj).Elem()
			objs = reflect.Append(objs, objValue)
		}
	}

	return objs.Interface(), nil
}
