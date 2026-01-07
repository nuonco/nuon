package provisionsandboxapplyplan

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queues/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queues/signal"
)

func init() {
	catalog.RegisterSignal(signal.SignalTypeInstallProvisionSandboxApplyPlan, func() signal.Signal {
		return &Signal{}
	})
}
