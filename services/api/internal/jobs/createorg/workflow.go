package createorg

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
)

type CreateOrgRequest struct {
	OrgID string `validate:"required"`
}

type CreateOrgResponse struct{}

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateOrg(ctx workflow.Context, req CreateOrgRequest) (CreateOrgResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create org")
	return CreateOrgResponse{}, nil
}
