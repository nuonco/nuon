package get

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

type Options struct {
	FieldTimeout time.Duration
	RootDir      string
}

type get struct {
	dst  any
	opts *Options
}

func Parse(ctx context.Context, obj any, opts *Options) error {
	v := validator.New()
	if err := v.StructCtx(ctx, opts); err != nil {
		return err
	}

	getter := &get{
		dst:  obj,
		opts: opts,
	}

	if err := getter.GetAll(ctx); err != nil {
		return err
	}

	return nil
}
