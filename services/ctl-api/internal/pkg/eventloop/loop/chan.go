package loop

import "go.temporal.io/sdk/workflow"

func (l *Loop[T, R]) drainSignals(ch workflow.ReceiveChannel) []T {
	var signals []T
	for {
		var signal T
		ok := ch.ReceiveAsync(&signal)
		if !ok {
			break
		}

		signals = append(signals, signal)
	}

	return signals
}
