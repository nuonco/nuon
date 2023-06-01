package callbacks

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline"
)

func Noop(context.Context, hclog.Logger, terminal.UI, []byte) error {
	return nil
}

func MapNoop(fn callbackNoop) pipeline.CallbackFn {
	return fn.callback
}

type callbackNoop func(context.Context) error

func (c callbackNoop) callback(ctx context.Context, log hclog.Logger, ui terminal.UI, byts []byte) error {
	if err := c(ctx); err != nil {
		return fmt.Errorf("unable to execute callback: %w", err)
	}
	return nil
}
