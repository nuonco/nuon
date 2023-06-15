package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline"
)

// execBytesLogFn is a function that just does an init, and returns bytes output
type execBytesLogFn func(context.Context, hclog.Logger) ([]byte, error)

func MapBytesLog(fn execBytesLogFn) pipeline.ExecFn {
	return fn.exec
}

func (p execBytesLogFn) exec(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]byte, error) {
	out, err := p(ctx, l)
	return out, err
}
