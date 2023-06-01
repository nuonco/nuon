package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline"
)

// execInitFn is a function that just does an init, and does not return output
type execInitFn func(context.Context) error

func MapInit(fn execInitFn) pipeline.ExecFn {
	return fn.exec
}

func (p execInitFn) exec(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]byte, error) {
	err := p(ctx)
	return nil, err
}
