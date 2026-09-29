package preflight

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

const probeTimeout = 10 * time.Second

func nopLogger() *zap.Logger { return zap.NewNop() }

func nopMetrics() metrics.Writer {
	mw, err := metrics.New(validator.New(),
		metrics.WithDisable(true),
		metrics.WithLogger(nopLogger()),
	)
	if err != nil {
		return nil
	}

	return mw
}
