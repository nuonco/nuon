package deployments

import (
	"context"

	"github.com/bufbuild/connect-go"
	deploymentv1 "github.com/powertoolsdev/protos/api/generated/types/deployment/v1"
)

func (s *server) CreateDeployment(
	ctx context.Context,
	req *connect.Request[deploymentv1.CreateDeploymentRequest],
) (*connect.Response[deploymentv1.CreateDeploymentResponse], error) {
	res := connect.NewResponse(&deploymentv1.CreateDeploymentResponse{})
	return res, nil
}
