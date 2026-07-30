package componenthealthnotify

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func init() {
	catalog.Register(ComponentUnhealthySignalType, func() signal.Signal {
		return &ComponentUnhealthySignal{}
	})
	catalog.Register(ComponentRecoveredSignalType, func() signal.Signal {
		return &ComponentRecoveredSignal{}
	})
	catalog.Register(InstallDegradedSignalType, func() signal.Signal {
		return &InstallDegradedSignal{}
	})
}
