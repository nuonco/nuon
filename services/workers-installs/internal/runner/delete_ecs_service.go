package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteServiceRequest struct {
	ClusterARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	Auth *credentials.Config `validate:"required"`
}

type DeleteServiceResponse struct{}

func (a *Activities) DeleteECSService(ctx context.Context, req *DeleteServiceRequest) (*DeleteServiceResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	ecsReq := &ecs.DeleteServiceInput{
		Cluster: generics.ToPtr(req.ClusterARN),
		Service: generics.ToPtr(fmt.Sprintf("waypoint-runner-%s", req.InstallID)),
		Force:   generics.ToPtr(true),
	}
	if _, err := ecsClient.DeleteService(ctx, ecsReq); err != nil {
		return nil, fmt.Errorf("unable to delete service: %w", err)
	}

	return &DeleteServiceResponse{}, nil
}
