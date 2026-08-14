package shutdownbeacon

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
)

func TestWriteAuditIgnoresUnavailableExporter(t *testing.T) {
	core, observed := observer.New(zap.WarnLevel)
	b := &Beacon{audit: new(audit.Writer), l: zap.New(core)}

	b.writeAudit()
	if observed.Len() != 0 {
		t.Fatalf("unavailable audit exporter produced %d warnings, want 0", observed.Len())
	}
}
