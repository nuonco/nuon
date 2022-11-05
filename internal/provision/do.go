package provision

import "context"

type DoRequest struct{}

type DoResponse struct{}

func (a *Activities) Do(ctx context.Context, req DoRequest) (DoResponse, error) {
	var resp DoResponse
	return resp, nil
}
