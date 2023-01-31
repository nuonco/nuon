package deployments

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	deploymentv1 "github.com/powertoolsdev/protos/api/generated/types/deployment/v1"
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
	deployments, _, err := s.Svc.GetComponentDeployments(ctx, req.Msg.InstallIds, &models.ConnectionOptions{})
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
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// TODO: add new service to retrieve deployments by app IDs
	deployments, _, err := s.Svc.GetComponentDeployments(ctx, req.Msg.AppIds, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get install deployments: %w", err)
	}

	return connect.NewResponse(&deploymentv1.GetDeploymentsByAppsResponse{
		Deployments: converters.DeploymentModelsToProtos(deployments),
	}), nil
}
