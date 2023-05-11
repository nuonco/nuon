package deployments

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	deploymentv1 "github.com/powertoolsdev/mono/pkg/types/api/deployment/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) GetDeployment(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentRequest],
) (*connect.Response[deploymentv1.GetDeploymentResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deployment, err := s.Svc.GetDeployment(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to get deployment: %w", err)
	}

	return connect.NewResponse(&deploymentv1.GetDeploymentResponse{
		Deployment: converters.DeploymentModelToProto(deployment),
	}), nil
}

func (s *server) GetDeploymentsByInstalls(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentsByInstallsRequest],
) (*connect.Response[deploymentv1.GetDeploymentsByInstallsResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// TODO: add new service to retrieve deployments by install IDs
	deployments, _, err := s.Svc.GetInstallDeployments(ctx, req.Msg.InstallIds, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get install deployments: %w", err)
	}

	return connect.NewResponse(&deploymentv1.GetDeploymentsByInstallsResponse{
		Deployments: converters.DeploymentModelsToProtos(deployments),
	}), nil
}

func (s *server) GetDeploymentsByComponents(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentsByComponentsRequest],
) (*connect.Response[deploymentv1.GetDeploymentsByComponentsResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	deployments, _, err := s.Svc.GetComponentDeployments(ctx, req.Msg.ComponentIds, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get component deployments: %w", err)
	}

	return connect.NewResponse(&deploymentv1.GetDeploymentsByComponentsResponse{
		Deployments: converters.DeploymentModelsToProtos(deployments),
	}), nil
}

func (s *server) GetDeploymentsByApps(
	ctx context.Context,
	req *connect.Request[deploymentv1.GetDeploymentsByAppsRequest],
) (*connect.Response[deploymentv1.GetDeploymentsByAppsResponse], error) {
	// run protobuf validations
	// TODO 174 temporarily disable validations until migration to shortIDs is complete
	// if err := req.Msg.Validate(); err != nil {
	// 	return nil, fmt.Errorf("input validation failed: %w", err)
	// }

	// TODO: add new service to retrieve deployments by app IDs
	deployments, _, err := s.Svc.GetAppDeployments(ctx, req.Msg.AppIds, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get install deployments: %w", err)
	}

	return connect.NewResponse(&deploymentv1.GetDeploymentsByAppsResponse{
		Deployments: converters.DeploymentModelsToProtos(deployments),
	}), nil
}
