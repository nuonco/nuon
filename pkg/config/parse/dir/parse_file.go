package dir

import (
	"bytes"
	"fmt"
	"io"
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

func (p *parser) parseFile(path, group string, obj any) (bool, error) {
	if !strings.HasSuffix(path, p.opts.Ext) {
		path = path + p.opts.Ext
	}

	exists, err := p.fs.Exists(path)
	if err != nil {
		return false, errors.Wrap(err, "unable to check that file exists")
	}
	if !exists {
		return false, nil
	}

	empty, err := afero.IsEmpty(p.fs, path)
	if err != nil {
		return false, errors.Wrap(err, "unable to check that file is empty")
	}
	if empty {
		return false, nil
	}

	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return false, errors.Wrap(err, "unable to read path")
	}

	if err := p.opts.ParserFn(io.NopCloser(bytes.NewReader(contents)), path, obj); err != nil {
		return false, ErrParseFile{
			Err:  err,
			Name: path,
		}
	}
	if p.opts.OnParsedFile != nil {
		if err := p.opts.OnParsedFile(ParsedFile{
			Path:     path,
			Group:    group,
			Contents: contents,
			Value:    obj,
		}); err != nil {
			return false, errors.Wrap(err, "unable to record parsed file")
		}
	}

	return true, nil
}
