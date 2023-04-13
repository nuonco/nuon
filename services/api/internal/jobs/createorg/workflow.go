package createorg

import "context"

type CreateOrgRequest struct {
	OrgID string `validate:"required"`
}

type CreateOrgResponse struct{}

type workflow struct{}

func (w *workflow) CreateOrg(ctx context.Context, req CreateOrgRequest) (CreateOrgResponse, error) {
	return CreateOrgResponse{}, nil
}
