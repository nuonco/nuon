package loop

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (l *Loop[SignalType, ReqSig]) ingestFn(ctx workflow.Context, req eventloop.EventLoopRequest, tags map[string]string) func(channel workflow.ReceiveChannel, _ bool) {
	log := l.logger(ctx)
	return func(channel workflow.ReceiveChannel, _ bool) {
		var signal SignalType
		channelOpen := channel.Receive(ctx, &signal)
		if !channelOpen {
			log.Info("channel was closed")
			return
		}

		// If a restart is already planned, drain incoming signals into l.pendingSignals for processing on re-entry.
		if l.restart || l.countExceeded {
			l.pendingSignals = append(l.pendingSignals, signal)
			return
		}

		// restart requires the signal to be handled on the fresh loop
		if signal.Restart() {
			l.pendingSignals = append(l.pendingSignals, signal)
			l.restart = true
			// l.restartChan.Close()
			return
		}

		l.sigcount++
		if l.sigcount > maxSignals {
			l.pendingSignals = append(l.pendingSignals, signal)
			l.countChan.Close()
			return
		}

		cg, has := l.conc[signal.ConcurrencyGroup()]
		if !has {
			cg = l.addGroup(signal.ConcurrencyGroup())
			l.concwg.Add(1)
			workflow.Go(ctx, l.spawnQueueHandler(req, cg, tags))
		}
		cg.queue = append(cg.queue, signal)
		// Force a state transition, causing all coroutine Await() callbacks to fire
		// NOTE(sdboyer) this would probably be smoother with temporal's builtin channels. But the buffered impl has concerning notes in
		// the comments about 'best effort' to be async, and we don't want any possibility that this main loop blocks
		workflow.SideEffect(ctx, func(ctx workflow.Context) any { return nil })
	}
}

func (l *Loop[SignalType, ReqSig]) spawnQueueHandler(req eventloop.EventLoopRequest, cg *concgroup[SignalType], tags map[string]string) func(ctx workflow.Context) {
	return func(ctx workflow.Context) {
		log := l.logger(ctx)
		defer l.concwg.Done()
		for {
			if len(cg.queue) == 0 {
				workflow.Await(ctx, func() bool {
					return len(cg.queue) > 0 || l.restart || l.countExceeded || l.notexist
				})
			}

			l.checkExists(ctx, req)
			if l.notexist || l.existsErr != nil {
				l.stopChan.Close()
				return
			}

			// Restarts or count exceeded means remaining signals will be picked up; put them in pending
			// and return
			if l.restart || l.countExceeded {
				// We prepend these to the pending list because anything that's already in there is either a)
				// from another concurrency group, in which case ordering does not matter, or b) was caught
				// by the signal ingestor after a restart was marked and put directly in pending, in which
				// case it definitely came after the signals in this queue.
				l.pendingSignals = append(cg.queue, l.pendingSignals...)
				return
			}

			signal := cg.queue[0]
			cg.queue = cg.queue[1:]

			if err := signal.Validate(l.V); err != nil {
				log.Error("invalid signal", zap.Error(err))
				continue
			}

			err := l.handleSignal(ctx, req, signal, tags)
			if err != nil {
				log.Error("error handling signal", zap.String("signaltype", string(signal.SignalType())), zap.Error(err))
				return
			}

			if signal.Stop() {
				l.stopChan.Close()
			}
		}
	}
}
