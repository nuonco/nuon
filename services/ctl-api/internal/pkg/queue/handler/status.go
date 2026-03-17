package handler

const (
	StatusQueryName string = "status"
	StatusQueryType        = handlerTypeQuery
)

type StatusRequest struct{}

type StatusResponse struct {
	Finished bool
	Canceled bool
}

// statusHandler is the query handler for the "status" query.
// Note: Temporal query handlers must NOT take workflow.Context as a parameter;
// the SDK's HandleQuery dispatches via executeFunction (not executeFunctionWithWorkflowContext),
// so injecting ctx would cause "reflect: Call with too few input arguments".
func (h *handler) statusHandler(req *StatusRequest) (*StatusResponse, error) {
	return &StatusResponse{
		Finished: h.finished,
		Canceled: h.canceled,
	}, nil
}
