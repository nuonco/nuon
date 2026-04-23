package callback

import (
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type Event string

const (
	OnEnqueue  Event = "on_enqueue"
	OnValidate Event = "on_validate"
	OnExecute  Event = "on_execute"
	OnSuccess  Event = "on_success"
	OnError    Event = "on_error"
)

// AllEvents returns all valid callback event types.
func AllEvents() []Event {
	return []Event{OnEnqueue, OnValidate, OnExecute, OnSuccess, OnError}
}

// CallbackRequest describes a callback to register when enqueuing a signal.
type CallbackRequest struct {
	Event         Event                  `validate:"required"`
	UpdateHandler signaldb.UpdateHandler `validate:"required"`
}

// CallbackPayload is the request payload sent as the update argument when invoking a callback.
type CallbackPayload struct {
	Event         Event  `json:"event"`
	QueueSignalID string `json:"queue_signal_id"`
	QueueID       string `json:"queue_id"`
	ErrMessage    string `json:"err_message,omitempty"`
}
