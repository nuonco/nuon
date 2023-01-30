package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	appv1 "github.com/powertoolsdev/protos/api/generated/types/app/v1"
)

func (s *server) DeleteApp(
	ctx context.Context,
	req *connect.Request[appv1.DeleteAppRequest],
) (*connect.Response[appv1.DeleteAppResponse], error) {
	// run protobuf validations
	if err := req.Msg.ValidateAll(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deleted, err := s.Svc.DeleteApp(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to delete app: %w", err)
	}

	return connect.NewResponse(&appv1.DeleteAppResponse{
		Deleted: deleted,
	}), nil
}
