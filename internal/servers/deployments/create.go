package deployments

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	deploymentv1 "github.com/powertoolsdev/protos/api/generated/types/deployment/v1"
)

func (s *server) CreateDeployment(
	ctx context.Context,
	req *connect.Request[deploymentv1.CreateDeploymentRequest],
) (*connect.Response[deploymentv1.CreateDeploymentResponse], error) {
	deployment, err := s.Svc.CreateDeployment(ctx, &models.DeploymentInput{
		ComponentID: req.Msg.ComponentId,
		CreatedByID: &req.Msg.CreatedById,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create deployment: %w", err)
	}

	return connect.NewResponse(&deploymentv1.CreateDeploymentResponse{
		Deployment: converters.DeploymentModelToProto(deployment),
	}), nil
}
