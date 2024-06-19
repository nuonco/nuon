package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteServiceRequest struct {
	IAMRoleARN string `validate:"required"`
	ClusterARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	TwoStepConfig *assumerole.TwoStepConfig `validate:"required"`
}

type DeleteServiceResponse struct{}

func (a *Activities) DeleteECSService(ctx context.Context, req DeleteServiceRequest) (*DeleteServiceResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.IAMRoleARN, req.Region, req.TwoStepConfig)
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
