package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
)

const (
	defaultRoleSessionName string = "noun-terraform-plugin"
)

// TODO(jm): convert this to a protocol buffer and/or return the docker
//
// NOTE: between this, and the mapper to map the image into a docker.Image, we should be able to support using this
// registry with the `docker-pull` build plugin, to sync images.
type AccessInfo struct {
	Image string
	Tag   string
	Auth  struct {
		Username      string
		Password      string
		ServerAddress string
	}
}

// AccessInfoFunc
func (r *Registry) AccessInfoFunc() interface{} {
	return r.AccessInfo
}

func (r *Registry) AccessInfo(ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	src *component.Source,
) (*AccessInfo, error) {
	authProvider, err := ecrauthorization.New(r.v,
		ecrauthorization.WithAssumeRoleArn(r.config.RoleARN),
		ecrauthorization.WithAssumeRoleSessionName(defaultRoleSessionName),
		ecrauthorization.WithRepository(r.config.Repository),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	accessInfo, err := r.getAccessInfo(ctx, authProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}

	// fetch ecr credentials here
	return accessInfo, nil
}

func (r *Registry) getAccessInfo(ctx context.Context, client ecrauthorization.Client) (*AccessInfo, error) {
	authorization, err := client.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization: %w", err)
	}

	return &AccessInfo{
		Image: r.config.Repository,
		Tag:   r.config.Tag,
		Auth: struct {
			Username      string
			Password      string
			ServerAddress string
		}{
			Username:      authorization.Username,
			Password:      authorization.RegistryToken,
			ServerAddress: authorization.ServerAddress,
		},
	}, nil
}
