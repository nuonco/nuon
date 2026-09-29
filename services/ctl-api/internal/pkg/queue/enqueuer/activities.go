package enqueuer

type Activities struct {
	e *Enqueuer
}

func NewActivities(e *Enqueuer) *Activities {
	return &Activities{e: e}
}
