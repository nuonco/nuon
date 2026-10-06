package dir

import (
	"context"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
)

type ParseOptions struct {
	Root         string                                 `validate:"required"`
	Ext          string                                 `validate:"required"`
	ParserFn     func(io.ReadCloser, string, any) error `validate:"required"`
	OnParsedFile func(ParsedFile) error

	// SkipDirs are name-tag directories whose files are not read.
	SkipDirs []string

	// IgnoreFileErrors are name-tag directories where a file that fails to
	// decode is left out and the rest of the directory is still loaded.
	IgnoreFileErrors []string
}

type ParsedFile struct {
	Path     string
	Group    string
	Contents []byte
	Value    any
}

type parser struct {
	fs   afero.Afero
	dst  any
	opts *ParseOptions
}

func Parse(ctx context.Context, fs afero.Fs, obj any, opts *ParseOptions) error {
	v := validator.New()
	if err := v.StructCtx(ctx, opts); err != nil {
		return err
	}

	parser := &parser{
		fs:   afero.Afero{fs},
		opts: opts,
		dst:  obj,
	}

	if err := parser.parse(ctx); err != nil {
		return err
	}

	return nil
}
