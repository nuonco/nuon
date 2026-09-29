package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/nuonco/nuon/pkg/pipeline"
)

type execInitFn func(context.Context) error

func MapInit(fn execInitFn) pipeline.ExecFn {
	return fn.exec
}

func (p execInitFn) exec(ctx context.Context, l hclog.Logger) ([]byte, error) {
	err := p(ctx)
	return nil, err
}
