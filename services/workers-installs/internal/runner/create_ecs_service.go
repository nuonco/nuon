package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateECSServiceRequest struct {
	ClusterARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	SecurityGroupID   string   `validate:"required"`
	SubnetIDs         []string `validate:"required"`
	TaskDefinitionARN string   `validate:"required"`

	Auth *credentials.Config `validate:"required"`
}

type CreateECSServiceResponse struct{}

func (a *Activities) CreateECSService(ctx context.Context, req *CreateECSServiceRequest) (*CreateECSServiceResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	svc, err := a.getECSService(ctx, ecsClient, req)
	if err != nil {
		return nil, fmt.Errorf("unable to describe services: %w", err)
	}

	if svc == nil {
		return a.createECSService(ctx, ecsClient, req)
	}

	return a.updateECSService(ctx, ecsClient, req)
}

func (a *Activities) getECSService(ctx context.Context, ecsClient *ecs.Client, req *CreateECSServiceRequest) (*ecstypes.Service, error) {
	inp := &ecs.DescribeServicesInput{
		Services: []string{
			fmt.Sprintf("waypoint-runner-%s", req.InstallID),
		},
		Cluster: generics.ToPtr(req.ClusterARN),
	}

	resp, err := ecsClient.DescribeServices(ctx, inp)
	if err != nil {
		return nil, fmt.Errorf("unable to describe services: %w", err)
	}

	if len(resp.Services) != 1 {
		return nil, nil
	}

	if *resp.Services[0].Status == "INACTIVE" {
		return nil, nil
	}

	return &resp.Services[0], nil
}

func (a *Activities) createECSService(ctx context.Context, ecsClient *ecs.Client, req *CreateECSServiceRequest) (*CreateECSServiceResponse, error) {
	ecsReq := &ecs.CreateServiceInput{
		Cluster:              generics.ToPtr(req.ClusterARN),
		DesiredCount:         generics.ToPtr(int32(1)),
		LaunchType:           ecstypes.LaunchTypeFargate,
		ClientToken:          generics.ToPtr(req.InstallID),
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

	_, err := ecsClient.CreateService(ctx, ecsReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create service: %w", err)
	}

	return &CreateECSServiceResponse{}, nil
}

func (a *Activities) updateECSService(ctx context.Context, ecsClient *ecs.Client, req *CreateECSServiceRequest) (*CreateECSServiceResponse, error) {
	ecsReq := &ecs.UpdateServiceInput{
		Cluster:              generics.ToPtr(req.ClusterARN),
		DesiredCount:         generics.ToPtr(int32(1)),
		Service:              generics.ToPtr(fmt.Sprintf("waypoint-runner-%s", req.InstallID)),
		EnableECSManagedTags: generics.ToPtr(true),
		TaskDefinition:       generics.ToPtr(req.TaskDefinitionARN),
		NetworkConfiguration: &ecstypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecstypes.AwsVpcConfiguration{
				Subnets:        req.SubnetIDs,
				SecurityGroups: []string{req.SecurityGroupID},
				AssignPublicIp: ecstypes.AssignPublicIpEnabled,
			},
		},
	}

	_, err := ecsClient.UpdateService(ctx, ecsReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create service: %w", err)
	}

	return &CreateECSServiceResponse{}, nil
}
