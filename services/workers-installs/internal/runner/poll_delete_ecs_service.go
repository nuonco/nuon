package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	defaultPollDeletionDuration time.Duration = time.Second * 10
)

type PollDeleteECSServiceRequest struct {
	ClusterARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	Auth *credentials.Config `validate:"required"`
}

type PollDeleteECSServiceResponse struct{}

func (a *Activities) PollDeleteService(ctx context.Context, req *PollDeleteECSServiceRequest) (*PollDeleteECSServiceResponse, error) {
	ecsClient, err := a.getECSClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client: %w", err)
	}

	ecsReq := &ecs.DescribeServicesInput{
		Cluster: generics.ToPtr(req.ClusterARN),
		Services: []string{
			fmt.Sprintf("waypoint-runner-%s", req.InstallID),
		},
	}

	for {
		resp, err := ecsClient.DescribeServices(ctx, ecsReq)
		if err != nil {
			return nil, fmt.Errorf("unable to describe services: %w", err)
		}
		if len(resp.Services) > 1 {
			return nil, fmt.Errorf("unexpected services returned: %w", err)
		}
		if len(resp.Services) < 1 {
			return nil, nil
		}

		if *resp.Services[0].Status == "INACTIVE" {
			return nil, nil
		}

		time.Sleep(defaultPollDeletionDuration)
	}

	return nil, nil
}
