package loop

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (l *Loop[SignalType, ReqSig]) checkExists(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	if l.ExistsHook != nil {
		log := l.logger(ctx)

		var exists bool
		exists, l.existsErr = l.ExistsHook(ctx, req)
		l.notexist = !exists
		if l.existsErr != nil {
			// This should only be reachable in the event of an underlying temporal error.
			log.Error("error checking for existence of underlying object", zap.Error(l.existsErr))
		}
	}
	return nil
}
