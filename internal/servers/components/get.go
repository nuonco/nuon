package components

import (
	"context"

	"github.com/bufbuild/connect-go"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) GetComponent(
	ctx context.Context,
	req *connect.Request[componentv1.GetComponentRequest],
) (*connect.Response[componentv1.GetComponentResponse], error) {
	res := connect.NewResponse(&componentv1.GetComponentResponse{})
	return res, nil
}

func (s *server) GetComponentsByApp(
	ctx context.Context,
	req *connect.Request[componentv1.GetComponentsByAppRequest],
) (*connect.Response[componentv1.GetComponentsByAppResponse], error) {
	res := connect.NewResponse(&componentv1.GetComponentsByAppResponse{})
	return res, nil
}
