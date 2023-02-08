package sync

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	assumerole "github.com/powertoolsdev/go-aws-assume-role"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint/configs"
)

// This builds a valid image source by fetching a token that can be used for docker authentication
// https://docs.aws.amazon.com/sdk-for-go/api/service/ecr/#ECR.GetAuthorizationToken to the org's ECR registry
func (p *planner) getSourceImage(ctx context.Context) (*configs.SyncImageSource, error) {
	assumer, err := assumerole.New(p.V,
		assumerole.WithRoleARN(p.OrgMetadata.IamRoleArns.InstancesRoleArn),
		assumerole.WithRoleSessionName("workers-executors-sync-image"))
	if err != nil {
		return nil, fmt.Errorf("unable to create role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to assume role: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	authData, err := p.getAuthorizationData(ctx, ecrClient)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization token: %w", err)
	}

	src, err := p.getSource(ctx, authData)
	if err != nil {
		return nil, fmt.Errorf("unable to get source image: %w", err)
	}

	return src, nil
}

func (p *planner) getSource(_ context.Context, data *ecr_types.AuthorizationData) (*configs.SyncImageSource, error) {
	auth, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("unable to decode auth string: %w", err)
	}

	authPieces := strings.SplitN(string(auth), ":", 2)

	return &configs.SyncImageSource{
		RegistryToken: authPieces[1],
		Username:      authPieces[0],
		ServerAddress: *data.ProxyEndpoint,
		Image:         fmt.Sprintf("766121324316.dkr.ecr.us-west-2.amazonaws.com/%s/%s", p.Metadata.OrgShortId, p.Metadata.AppShortId),
		Tag:           p.Metadata.DeploymentShortId,
	}, nil
}

type awsECRClient interface {
	GetAuthorizationToken(context.Context, *ecr.GetAuthorizationTokenInput, ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

func (p *planner) getAuthorizationData(ctx context.Context, client awsECRClient) (*ecr_types.AuthorizationData, error) {
	params := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{
			p.OrgMetadata.EcrRegistryId,
		},
	}

	resp, err := client.GetAuthorizationToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization token: %w", err)
	}

	return &resp.AuthorizationData[0], nil
}
