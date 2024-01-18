package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateECSServiceRequest struct {
	IAMRoleARN string `validate:"required"`
	ClusterARN string `validate:"required"`
	InstallID  string `validate:"required"`

	SecurityGroupID   string   `validate:"required"`
	SubnetIDs         []string `validate:"required"`
	TaskDefinitionARN string   `validate:"required"`
}

type CreateECSServiceResponse struct{}

func (a *Activities) CreateECSService(ctx context.Context, req *CreateECSServiceRequest) (*CreateECSServiceResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.IAMRoleARN)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	ecsReq := &ecs.CreateServiceInput{
		Cluster:              generics.ToPtr(req.ClusterARN),
		DesiredCount:         generics.ToPtr(int32(1)),
		LaunchType:           ecstypes.LaunchTypeFargate,
		ServiceName:          generics.ToPtr(fmt.Sprintf("waypoint-runner-%s", req.InstallID)),
		EnableECSManagedTags: true,
		TaskDefinition:       generics.ToPtr(req.TaskDefinitionARN),
		NetworkConfiguration: &ecstypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecstypes.AwsVpcConfiguration{
				Subnets:        req.SubnetIDs,
				SecurityGroups: []string{req.SecurityGroupID},
				AssignPublicIp: ecstypes.AssignPublicIpEnabled,
			},
		},
		Tags: []ecstypes.Tag{
			{
				Key:   aws.String(defaultRunnerTagName),
				Value: aws.String(defaultRunnerTagValue),
			},
			{
				Key:   generics.ToPtr(defaultRunnerIDTagName),
				Value: generics.ToPtr(req.InstallID),
			},
		},
	}

	if _, err := ecsClient.CreateService(ctx, ecsReq); err != nil {
		return nil, fmt.Errorf("unable to create ecs service: %w", err)
	}

	return &CreateECSServiceResponse{}, nil
}
