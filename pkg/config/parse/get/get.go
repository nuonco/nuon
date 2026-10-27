package get

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

type Options struct {
	FieldTimeout time.Duration
	RootDir      string
	OnLocalFile  func(path string, contents []byte) error

	// WorkDir, relative to RootDir, is where relative paths resolve from when
	// no config section sets its own subdirectory.
	WorkDir string

	// Restricted is for configs from an untrusted source parsed on the control
	// plane: local reads are confined to RootDir and remote fetches to the
	// public internet.
	Restricted bool
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
