package publish

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "customer_managed_bundle_publish"

func init() {
	catalog.Register(SignalType, func() signal.Signal { return &Signal{} })
}
