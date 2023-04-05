package provision

import "context"

type Activities struct{}

func NewActivities() *Activities {
	return &Activities{}
}

type NoopRequest struct{}
type NoopResponse struct{}

func (a *Activities) Noop(ctx context.Context, req NoopRequest) (NoopResponse, error) {
	return NoopResponse{}, nil
}
