package signal

import "time"

// DeriveMaxInFlightAge returns the configured max-in-flight age for a signal,
// or 0 if the signal does not opt in. Callers should treat 0 as "dedup forever"
// (the legacy behavior).
func DeriveMaxInFlightAge(sig Signal) time.Duration {
	if t, ok := sig.(SignalWithMaxInFlightAge); ok {
		return t.MaxInFlightAge()
	}
	return 0
}
