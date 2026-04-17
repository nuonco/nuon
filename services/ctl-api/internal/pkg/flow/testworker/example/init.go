package example

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func init() {
	catalog.Register(FakeGenerateStepsSignalType, func() signal.Signal {
		return &FakeGenerateStepsSignal{}
	})
	catalog.Register(FakeStepSignalType, func() signal.Signal {
		return &FakeStepSignal{}
	})
}
