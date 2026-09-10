package testworker

import (
	"os"
	"time"
)

// NUON_FLOW_PHASE_LEDGER=1 enables per-test phase timing, emitted via t.Logf
// so it lands in verbose test output next to each test's wall time.
var phaseLedgerEnabled = os.Getenv("NUON_FLOW_PHASE_LEDGER") == "1"

// Phase marks live on the FlowTestSuite value, which is one per running case —
// cases execute concurrently, so nothing here may be package-global.
func (e *FlowTestSuite) phase(name string) {
	if !phaseLedgerEnabled {
		return
	}
	now := time.Now()
	if !e.ledgerOn {
		e.ledgerOn = true
		e.ledgerStart = now
		e.ledgerMark = now
	}
	e.T().Logf("PHASE %q +%s (total %s)", name,
		now.Sub(e.ledgerMark).Round(time.Millisecond), now.Sub(e.ledgerStart).Round(time.Millisecond))
	e.ledgerMark = now
}
