package signaldb

import (
	"encoding/json"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
