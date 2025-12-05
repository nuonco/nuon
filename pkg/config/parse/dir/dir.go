package dir

import (
	"context"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
)

type ParseOptions struct {
	Root     string                                 `validate:"required"`
	Ext      string                                 `validate:"required"`
	ParserFn func(io.ReadCloser, string, any) error `validate:"required"`
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
