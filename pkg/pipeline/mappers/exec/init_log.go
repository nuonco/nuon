package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/nuonco/nuon/pkg/pipeline"
)

type execInitLogFn func(context.Context, hclog.Logger) error

func MapInitLog(fn execInitLogFn) pipeline.ExecFn {
	return fn.exec
}

func (p execInitLogFn) exec(ctx context.Context, l hclog.Logger) ([]byte, error) {
	err := p(ctx, l)
	return nil, err
}
