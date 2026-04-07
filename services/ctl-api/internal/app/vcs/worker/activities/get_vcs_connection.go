package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetVCSConnectionRequest struct {
	VCSConnectionID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field VCSConnectionID
func (a *Activities) GetVCSConnection(ctx context.Context, req GetVCSConnectionRequest) (*app.VCSConnection, error) {
	var vcsConn app.VCSConnection
	res := a.db.WithContext(ctx).
		First(&vcsConn, "id = ?", req.VCSConnectionID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", res.Error)
	}
	return &vcsConn, nil
}
