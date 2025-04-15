package dir

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type ErrEmptyFile struct {
	Name string
}

func (e ErrEmptyFile) Error() string {
	return fmt.Sprintf("required file %s was empty", e.Name)
}

type ErrMissingFile struct {
	Name string
}

func (e ErrMissingFile) Error() string {
	return fmt.Sprintf("required file %s was not found", e.Name)
}

type ErrMissingDir struct {
	Name string
}

func (e ErrMissingDir) Error() string {
	return fmt.Sprintf("required directory %s was not found", e.Name)
}

type ErrParseFile struct {
	Err  error
	Name string
}

func (e ErrParseFile) Error() string {
	return fmt.Sprintf("unable to parse %s", e.Name)
}

func (e ErrParseFile) Unwrap() error {
	return e.Err
}

func (p *parser) parseFile(path string, obj any, required, nonempty bool) error {
	if !strings.HasSuffix(path, p.opts.Ext) {
		path = path + p.opts.Ext
	}

	exists, err := p.fs.Exists(path)
	if err != nil {
		return errors.Wrap(err, "unable to check that file exists")
	}
	if !exists {
		if required {
			return ErrMissingFile{
				Name: path,
			}
		}

		return nil
	}

	empty, err := afero.IsEmpty(p.fs, path)
	if err != nil {
		return errors.Wrap(err, "unable to check that file is empty")
	}
	if empty {
		if nonempty {
			return ErrEmptyFile{
				Name: path,
			}
		}

		return nil
	}

	fh, err := p.fs.Open(path)
	if err != nil {
		return errors.Wrap(err, "unable to open path")
	}

	if err := p.opts.ParserFn(fh, path, obj); err != nil {
		return ErrParseFile{
			Err:  err,
			Name: path,
		}
	}

	return nil
}
