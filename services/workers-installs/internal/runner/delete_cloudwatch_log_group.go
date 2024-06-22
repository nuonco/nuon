package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteCloudwatchLogGroupRequest struct {
	LogGroupName string `validate:"required"`
	Region       string `validate:"required"`

	Auth *credentials.Config `validate:"required"`
}

type DeleteCloudwatchLogGroupResponse struct{}

func (a *Activities) DeleteCloudwatchLogGroup(ctx context.Context, req *DeleteCloudwatchLogGroupRequest) (*DeleteCloudwatchLogGroupResponse, error) {
	cwClient, err := a.getCloudwatchClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	_, err = cwClient.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: generics.ToPtr(req.LogGroupName),
	})
	if err != nil {
		nfErr := &cloudwatchlogstypes.ResourceNotFoundException{}
		if errors.As(err, &nfErr); err != nil {
			return &DeleteCloudwatchLogGroupResponse{}, nil
		}

		return nil, fmt.Errorf("unable to create log group: %w", err)
	}

	return &DeleteCloudwatchLogGroupResponse{}, nil
}
