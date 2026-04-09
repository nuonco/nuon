package queue

const (
	ReadyHandlerName string = "ready"
	ReadyHandlerType        = handlerTypeQuery
)

type ReadyResponse struct {
	Ready bool
}

func (w *queue) readyHandler() (*ReadyResponse, error) {
	return &ReadyResponse{Ready: w.ready}, nil
}
