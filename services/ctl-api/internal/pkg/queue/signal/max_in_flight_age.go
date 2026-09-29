package signal

import "time"

func DeriveMaxInFlightAge(sig Signal) time.Duration {
	if t, ok := sig.(SignalWithMaxInFlightAge); ok {
		return t.MaxInFlightAge()
	}
	return 0
}
