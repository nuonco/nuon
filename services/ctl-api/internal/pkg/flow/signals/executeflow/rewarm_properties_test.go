package executeflow

import (
	"testing"

	"hegel.dev/go/hegel"
)

// The auto-rewarm gate must only arm on mutating updates: a terminal resident
// host re-warmed by read-only polls (poll-next-step, is-retryable) must never
// re-drive the conductor, and must decline so the Handler can finish cleanly.
func TestReadOnlyUpdatesNeverArmAutoRewarm(t *testing.T) {
	t.Run("ready iff a mutating update started; declined iff only reads ran and settled", hegel.Case(func(ht *hegel.T) {
		opCount := hegel.Draw(ht, hegel.Integers(1, 30))

		sig := Signal{}
		mutatingStarted := 0
		totalStarted := 0
		var inFlightEnds []func()

		for range opCount {
			op := hegel.Draw(ht, hegel.Integers(0, 2))
			switch op {
			case 0:
				inFlightEnds = append(inFlightEnds, sig.beginReadOnlyUpdate())
				totalStarted++
			case 1:
				inFlightEnds = append(inFlightEnds, sig.beginUpdate())
				mutatingStarted++
				totalStarted++
			case 2:
				if len(inFlightEnds) > 0 {
					inFlightEnds[len(inFlightEnds)-1]()
					inFlightEnds = inFlightEnds[:len(inFlightEnds)-1]
				}
			}

			if (mutatingStarted > 0) != sig.AutoExecuteReady() {
				ht.Fatalf("AutoExecuteReady=%v with %d mutating updates started", sig.AutoExecuteReady(), mutatingStarted)
			}
			if sig.AutoExecuteReady() && sig.AutoExecuteDeclined() {
				ht.Fatalf("AutoExecuteReady and AutoExecuteDeclined must be mutually exclusive")
			}
			if mutatingStarted == 0 && len(inFlightEnds) > 0 && sig.AutoExecuteDeclined() {
				ht.Fatalf("declined while a read-only update is still in flight")
			}
		}

		for _, end := range inFlightEnds {
			end()
		}

		if mutatingStarted > 0 {
			if !sig.AutoExecuteReady() {
				ht.Fatalf("mutating updates ran but host did not arm rewarm")
			}
			if sig.AutoExecuteDeclined() {
				ht.Fatalf("host declined despite mutating updates")
			}
		} else {
			if sig.AutoExecuteReady() {
				ht.Fatalf("read-only updates armed rewarm")
			}
			if totalStarted > 0 && !sig.AutoExecuteDeclined() {
				ht.Fatalf("settled read-only host did not decline")
			}
			if totalStarted == 0 && sig.AutoExecuteDeclined() {
				ht.Fatalf("host declined before any update ran")
			}
		}
	}))
}
