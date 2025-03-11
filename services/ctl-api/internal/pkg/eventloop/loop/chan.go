package loop

import "go.temporal.io/sdk/workflow"

func (l *Loop[SignalType, ReqSig]) drainSignals(ch workflow.ReceiveChannel) []SignalType {
	var signals []SignalType
	for {
		var signal SignalType
		ok := ch.ReceiveAsync(&signal)
		if !ok {
			break
		}

		signals = append(signals, signal)
	}

	return signals
}
