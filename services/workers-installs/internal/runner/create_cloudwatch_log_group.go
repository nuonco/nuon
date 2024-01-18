package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateCloudwatchLogGroupRequest struct {
	IAMRoleARN   string `validate:"required"`
	LogGroupName string `validate:"required"`
}

type CreateCloudwatchLogGroupResponse struct {
	LogGroupName string
}

func (a *Activities) CreateCloudwatchLogGroup(ctx context.Context, req *CreateCloudwatchLogGroupRequest) (*CreateCloudwatchLogGroupResponse, error) {
	cwClient, err := a.getCloudwatchClient(ctx, req.IAMRoleARN)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	_, err = cwClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: generics.ToPtr(req.LogGroupName),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create log group: %w", err)
	}

	return &CreateCloudwatchLogGroupResponse{
		LogGroupName: req.LogGroupName,
	}, nil
}
