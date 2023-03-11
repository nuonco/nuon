package components

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/api/component/v1"
)

func (s *server) DeleteComponent(
	ctx context.Context,
	req *connect.Request[componentv1.DeleteComponentRequest],
) (*connect.Response[componentv1.DeleteComponentResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deleted, err := s.Svc.DeleteComponent(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to delete component: %w", err)
	}

	return connect.NewResponse(&componentv1.DeleteComponentResponse{
		Deleted: deleted,
	}), nil
}
