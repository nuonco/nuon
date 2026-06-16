package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
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
		return nil, generics.TemporalGormError(res.Error, "unable to get vcs connection")
	}
	return &vcsConn, nil
}
