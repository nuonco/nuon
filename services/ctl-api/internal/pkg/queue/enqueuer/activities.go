package enqueuer

// Activities wraps the Enqueuer methods that should be registered as Temporal
// activities. This avoids exposing non-activity methods (e.g. Send) which
// would cause Temporal to panic on registration.
type Activities struct {
	e *Enqueuer
}

func NewActivities(e *Enqueuer) *Activities {
	return &Activities{e: e}
}
