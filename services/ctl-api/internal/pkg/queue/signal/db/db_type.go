package signaldb

import (
	"encoding/json"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

// signalJSON is the actual signal data that is used to store a signal
type signalJSON struct {
	Type signal.SignalType `json:"type"`
	Data signal.Signal     `json:"data"`
}

type anyJSON struct {
	Type signal.SignalType `json:"type"`
	Data json.RawMessage   `json:"data"`
}
