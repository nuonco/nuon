package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/powertoolsdev/mono/pkg/pipeline"
)

// execInitLogFn is a function that just does an init, and does not return output
type execInitLogFn func(context.Context, hclog.Logger) error

func MapInitLog(fn execInitLogFn) pipeline.ExecFn {
	return fn.exec
}

func (p execInitLogFn) exec(ctx context.Context, l hclog.Logger) ([]byte, error) {
	err := p(ctx, l)
	return nil, err
}
