package activities

import (
	"context"
)

type CancelRequest struct {
	ID        string `validate:"required"`
	Namespace string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) PkgSignalsCancel(ctx context.Context, req *CancelRequest) error {
	return a.evClient.Cancel(ctx, req.Namespace, req.ID)
}
