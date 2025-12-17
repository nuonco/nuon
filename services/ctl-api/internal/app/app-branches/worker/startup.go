package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
)

func (w *Workflows) startup(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	return nil
}
