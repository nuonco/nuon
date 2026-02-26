package preflight

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

// probeTimeout bounds a single check so one unreachable dependency cannot stall
// the whole run.
const probeTimeout = 10 * time.Second

// Preflight runs outside the fx graph — it has to work when config is broken
// enough that fx boot would fail — so the shared constructors it reuses get
// throwaway no-op dependencies instead of the real logger and metrics writer.

func nopLogger() *zap.Logger { return zap.NewNop() }

// WithLogger is not optional here: metrics.New otherwise builds its own
// development logger, whose debug lines would interleave with the results table.
func nopMetrics() metrics.Writer {
	mw, err := metrics.New(validator.New(),
		metrics.WithDisable(true),
		metrics.WithLogger(nopLogger()),
	)
	if err != nil {
		// Neither option can fail validation, so this is unreachable.
		return nil
	}

	return mw
}
