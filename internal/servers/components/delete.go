package components

import (
	"context"

	"github.com/bufbuild/connect-go"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) DeleteComponent(
	ctx context.Context,
	req *connect.Request[componentv1.DeleteComponentRequest],
) (*connect.Response[componentv1.DeleteComponentResponse], error) {
	res := connect.NewResponse(&componentv1.DeleteComponentResponse{})
	return res, nil
}
