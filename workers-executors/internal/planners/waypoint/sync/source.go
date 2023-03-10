package sync

import (
	"context"
	"fmt"

	awsecrauthorization "github.com/powertoolsdev/workers-executors/internal/aws-ecr-authorization"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint/configs"
)

const (
	defaultAssumeRoleSessionName string = "workers-executors-sync-image"
)

func (p *planner) getImageSource(ctx context.Context) (*configs.PrivateImageSource, error) {
	ecr, err := awsecrauthorization.New(p.V,
		awsecrauthorization.WithAssumeRoleArn(p.OrgMetadata.IamRoleArns.InstancesRoleArn),
		awsecrauthorization.WithAssumeRoleSessionName(defaultAssumeRoleSessionName),
		awsecrauthorization.WithRegistryID(p.OrgMetadata.EcrRegistryId),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecrauthorizer for private docker pull build: %w", err)
	}

	ecrAuth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &configs.PrivateImageSource{
		RegistryToken: ecrAuth.RegistryToken,
		Username:      ecrAuth.Username,
		ServerAddress: ecrAuth.ServerAddress,
		Image:         fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s/%s", p.OrgMetadata.EcrRegistryId, p.OrgMetadata.EcrRegion, p.Metadata.OrgShortId, p.Metadata.AppShortId),
		Tag:           p.Metadata.DeploymentShortId,
	}, nil
}
