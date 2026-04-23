package callback

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

// CallbackPayload is the request payload sent as the update argument when invoking a callback.
type CallbackPayload struct {
	Event         Event  `json:"event"`
	QueueSignalID string `json:"queue_signal_id"`
	QueueID       string `json:"queue_id"`
}
