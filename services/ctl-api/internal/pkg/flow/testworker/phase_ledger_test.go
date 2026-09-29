package testworker

import (
	"os"
	"time"
)

var phaseLedgerEnabled = os.Getenv("NUON_FLOW_PHASE_LEDGER") == "1"

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
