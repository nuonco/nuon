package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
)

type GetEventLoopsRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) GetEventLoops(ctx context.Context, req GetEventLoopsRequest) ([]bulk.EventLoop, error) {
	return a.helpers.GetEventLoops(ctx, req.OrgID)
}
