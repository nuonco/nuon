package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateECSTaskDefinitionRequest struct {
	IAMRoleARN string `validate:"required"`
	InstallID  string `validate:"required"`

	RunnerRoleARN string            `validate:"required"`
	EnvVars       map[string]string `validate:"required"`
	LogGroupName  string            `validate:"required"`
	Region        string            `validate:"required"`
	ServerCookie  string            `validate:"required"`
	AccessPointID string            `validate:"required"`
	FileSystemID  string            `validate:"required"`

	Args []string

	TwoStepConfig *assumerole.TwoStepConfig
}

type CreateECSTaskDefinitionResponse struct {
	TaskDefinitionARN string
}

func (a *Activities) CreateECSTaskDefinition(ctx context.Context, req *CreateECSTaskDefinitionRequest) (*CreateECSTaskDefinitionResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.IAMRoleARN, req.Region, req.TwoStepConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	taskDef := a.getTaskDefinition(req)

	taskDefResp, err := ecsClient.RegisterTaskDefinition(ctx, taskDef)
	if err != nil {
		return nil, fmt.Errorf("unable to register task definition: %w", err)
	}

	return &CreateECSTaskDefinitionResponse{
		TaskDefinitionARN: *taskDefResp.TaskDefinition.TaskDefinitionArn,
	}, nil
}

func (a *Activities) getLoggingOptions(req *CreateECSTaskDefinitionRequest) map[string]string {
	result := map[string]string{
		"awslogs-region":        req.Region,
		"awslogs-group":         req.LogGroupName,
		"awslogs-stream-prefix": fmt.Sprintf("waypoint-runner-%d", time.Now().Nanosecond()),
	}

	return result
}

func (a *Activities) getTaskDefinition(req *CreateECSTaskDefinitionRequest) *ecs.RegisterTaskDefinitionInput {
	env := make([]ecstypes.KeyValuePair, 0)
	for k, v := range req.EnvVars {
		env = append(env, ecstypes.KeyValuePair{
			Name:  generics.ToPtr(k),
			Value: generics.ToPtr(v),
		})
	}

	def := ecstypes.ContainerDefinition{
		Essential: generics.ToPtr(true),
		Command: append([]string{
			"runner",
			"agent",
			"-id=" + req.InstallID,
			"-liveness-tcp-addr=:1234",
			"-cookie=" + req.ServerCookie,
			"-state-dir=/data/runner",
			"-vv",
		}, req.Args...),
		Name:        generics.ToPtr("runner"),
		Image:       generics.ToPtr("public.ecr.aws/p7e3r5y0/waypoint:v0.1.0"),
		Environment: env,
		LogConfiguration: &ecstypes.LogConfiguration{
			LogDriver: ecstypes.LogDriverAwslogs,
			Options:   a.getLoggingOptions(req),
		},
		MountPoints: []ecstypes.MountPoint{
			{
				ContainerPath: generics.ToPtr("/data/runner"),
				ReadOnly:      generics.ToPtr(false),
				SourceVolume:  generics.ToPtr(defaultRunnerTagName),
			},
		},
	}

	return &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []ecstypes.ContainerDefinition{def},
		ExecutionRoleArn:     generics.ToPtr(req.RunnerRoleARN),
		Cpu:                  generics.ToPtr("512"),
		Memory:               generics.ToPtr("2048"),
		Family:               generics.ToPtr(defaultRunnerTagName),
		TaskRoleArn:          generics.ToPtr(req.RunnerRoleARN),
		NetworkMode:          ecstypes.NetworkModeAwsvpc,
		RequiresCompatibilities: []ecstypes.Compatibility{
			ecstypes.CompatibilityFargate,
		},
		Tags: []ecstypes.Tag{
			{
				Key:   aws.String(defaultRunnerTagName),
				Value: aws.String(defaultRunnerTagValue),
			},
			{
				Key:   aws.String("runner-id"),
				Value: aws.String(req.InstallID),
			},
		},
		Volumes: []ecstypes.Volume{
			{
				EfsVolumeConfiguration: &ecstypes.EFSVolumeConfiguration{
					AuthorizationConfig: &ecstypes.EFSAuthorizationConfig{
						AccessPointId: generics.ToPtr(req.AccessPointID),
					},
					FileSystemId:      generics.ToPtr(req.FileSystemID),
					TransitEncryption: ecstypes.EFSTransitEncryptionEnabled,
				},
				Name: aws.String(defaultRunnerTagName),
			},
		},
	}
}
