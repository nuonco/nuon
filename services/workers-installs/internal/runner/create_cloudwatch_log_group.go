package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateCloudwatchLogGroupRequest struct {
	IAMRoleARN   string `validate:"required"`
	LogGroupName string `validate:"required"`
	Region       string `validate:"required"`
}

type CreateCloudwatchLogGroupResponse struct {
	LogGroupName string
}

func (a *Activities) CreateCloudwatchLogGroup(ctx context.Context, req *CreateCloudwatchLogGroupRequest) (*CreateCloudwatchLogGroupResponse, error) {
	cwClient, err := a.getCloudwatchClient(ctx, req.IAMRoleARN, req.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	_, err = cwClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: generics.ToPtr(req.LogGroupName),
	})
	if err != nil {
		alreadyExistsErr := &cloudwatchlogstypes.ConflictException{}
		if errors.As(err, &alreadyExistsErr); err != nil {
			return &CreateCloudwatchLogGroupResponse{
				LogGroupName: req.LogGroupName,
			}, nil
		}

		return nil, fmt.Errorf("unable to create log group: %w", err)
	}

	return &CreateCloudwatchLogGroupResponse{
		LogGroupName: req.LogGroupName,
	}, nil
}
