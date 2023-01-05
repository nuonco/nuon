package deployments

import (
	"context"

	"github.com/bufbuild/connect-go"
	deploymentv1 "github.com/powertoolsdev/protos/api/generated/types/deployment/v1"
)

func (s *server) GetDeployment(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentRequest],
) (*connect.Response[deploymentv1.GetDeploymentResponse], error) {
	res := connect.NewResponse(&deploymentv1.GetDeploymentResponse{})
	return res, nil
}

func (s *server) GetDeploymentsByComponents(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentsByComponentsRequest],
) (*connect.Response[deploymentv1.GetDeploymentsByComponentsResponse], error) {
	res := connect.NewResponse(&deploymentv1.GetDeploymentsByComponentsResponse{})
	return res, nil
}
