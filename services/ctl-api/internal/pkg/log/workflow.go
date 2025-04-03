package log

import (
	"context"
	"runtime"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// WorkflowLogger returns a (*zap.Logger) that logs to both the log stream (if it is set in the context) and the
// underlying temporal logger.
func WorkflowLogger(ctx workflow.Context, attrs ...map[string]string) (*zap.Logger, error) {
	wfl := temporalzap.GetWorkflowLogger(ctx)

	ls, err := cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return wfl, nil
	}

	lp, err := NewOTELProvider(ls)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get otel provider")
	}

	l, err := NewLogStreamLogger(ls, lp, wfl, attrs...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create logger")
	}

	runtime.SetFinalizer(l, func(_ any) {
		cleanupCtx := context.Background()
		cleanupCtx, cancel := context.WithTimeout(cleanupCtx, time.Second)
		defer cancel()

		lp.ForceFlush(cleanupCtx)
		lp.Shutdown(cleanupCtx)
	})

	return l, nil
}
