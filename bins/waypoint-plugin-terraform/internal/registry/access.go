package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

const (
	defaultRoleSessionName string = "noun-terraform-plugin"
)

// AccessInfoFunc
func (r *Registry) AccessInfoFunc() interface{} {
	return r.AccessInfo
}

func (r *Registry) AccessInfo(ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	src *component.Source,
) (*terraformv1.AccessInfo, error) {
	authProvider, err := ecrauthorization.New(r.v,
		ecrauthorization.WithCredentials(r.config.Auth),
		ecrauthorization.WithRepository(r.config.Repository),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	accessInfo, err := r.getAccessInfo(ctx, authProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}

	return accessInfo, nil
}

func (r *Registry) getAccessInfo(ctx context.Context, client ecrauthorization.Client) (*terraformv1.AccessInfo, error) {
	authorization, err := client.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization: %w", err)
	}

	return &terraformv1.AccessInfo{
		Image: r.config.Repository,
		Tag:   r.config.Tag,
		Auth: &terraformv1.Auth{
			Username:      authorization.Username,
			Password:      authorization.RegistryToken,
			ServerAddress: authorization.ServerAddress,
		},
	}, nil
}
