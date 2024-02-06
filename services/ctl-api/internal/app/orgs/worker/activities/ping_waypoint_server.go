package activities

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PingWaypointServerRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) PingWaypointServer(ctx context.Context, req PingWaypointServerRequest) error {
	_, err := a.wpClient.GetVersionInfo(ctx, req.OrgID, &emptypb.Empty{})
	if err != nil {
		return err
	}

	return nil
}
