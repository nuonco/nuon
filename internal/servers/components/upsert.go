package components

import (
	"context"

	"github.com/bufbuild/connect-go"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) UpsertComponent(
	ctx context.Context,
	req *connect.Request[componentv1.UpsertComponentRequest],
) (*connect.Response[componentv1.UpsertComponentResponse], error) {
	res := connect.NewResponse(&componentv1.UpsertComponentResponse{})
	return res, nil
}
