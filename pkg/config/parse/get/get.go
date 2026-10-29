package get

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/publicnet"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

type Options struct {
	FieldTimeout time.Duration
	RootDir      string
	// Source is borrowed; when set, it is the authority for local file reads.
	Source      *sourcefs.FS
	OnLocalFile func(path string, contents []byte) error
}

type get struct {
	dst        any
	opts       *Options
	httpClient *http.Client
}

func Parse(ctx context.Context, obj any, opts *Options) error {
	v := validator.New()
	if err := v.StructCtx(ctx, opts); err != nil {
		return err
	}
	options := *opts
	if options.Source != nil && options.RootDir == "" {
		options.RootDir = options.Source.Root()
	}

	getter := &get{
		dst:  obj,
		opts: &options,
	}
	if options.Source != nil {
		getter.httpClient = publicnet.NewHTTPClient()
		defer getter.httpClient.CloseIdleConnections()
	}

	if err := getter.GetAll(ctx); err != nil {
		return err
	}

	return nil
}
