package loop

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (l *Loop[SignalType, ReqSig]) createSelector(ctx workflow.Context, req eventloop.EventLoopRequest, tags map[string]string) workflow.Selector {
	selector := workflow.NewSelector(ctx)

	signalChan := workflow.GetSignalChannel(ctx, req.ID)
	selector.AddReceive(signalChan, l.ingestFn(ctx, req, tags))

	l.stopChan = workflow.NewChannel(ctx)
	selector.AddReceive(l.stopChan, func(channel workflow.ReceiveChannel, _ bool) {
		l.stop = true
	})
	l.restartChan = workflow.NewChannel(ctx)
	selector.AddReceive(l.restartChan, func(channel workflow.ReceiveChannel, _ bool) {
		l.restart = true
	})
	l.countChan = workflow.NewChannel(ctx)
	selector.AddReceive(l.countChan, func(channel workflow.ReceiveChannel, _ bool) {
		l.countExceeded = true
	})

	var existsFut func(f workflow.Future)
	existsFut = func(f workflow.Future) {
		l.checkExists(ctx, req)
		selector.AddFuture(workflow.NewTimer(ctx, checkExistsInterval), existsFut)
	}
	selector.AddFuture(workflow.NewTimer(ctx, checkExistsInterval), existsFut)

	return selector
}

func (l *Loop[SignalType, ReqSig]) handlePending(ctx workflow.Context, req eventloop.EventLoopRequest, tags map[string]string, incomingPendingSignals []SignalType) {
	l.MW.Gauge(ctx, "event_loop.pending_signals", float64(len(incomingPendingSignals)), metrics.ToTags(tags)...)
	if len(incomingPendingSignals) > 0 {
		for _, pendingSignal := range incomingPendingSignals {
			cg, has := l.conc[pendingSignal.ConcurrencyGroup()]
			if !has {
				cg = l.addGroup(pendingSignal.ConcurrencyGroup())
				l.concwg.Add(1)
				workflow.Go(ctx, l.spawnQueueHandler(req, cg, tags))
			}
			// Don't count pending signals against the maxSignals total count. Otherwise, a large influx of signals (> maxSignals)
			// could cause the eventloop to continuously restart without ever processing any.
			cg.queue = append(cg.queue, pendingSignal)
		}

		// Give all the corroutines a kick
		workflow.SideEffect(ctx, func(ctx workflow.Context) any { return nil })
	}
}
