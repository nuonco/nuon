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

func (r *Registry) PushFunc() interface{} {
	return r.Push
}

func (r *Registry) Push(
	ctx context.Context,
	log hclog.Logger,
	bld *terraformv1.BuildOutput,
	ui terminal.UI,
	src *component.Source,
) (*terraformv1.Artifact, error) {
	u := ui.Status()
	defer u.Close()

	u.Update("initialized push")
	store, err := r.getStore()
	if err != nil {
		u.Step(terminal.StatusError, "unable to get store")
		return nil, fmt.Errorf("unable to get store: %w", err)
	}
	u.Step(terminal.StatusOK, "successfully opened store")

	authProvider, err := ecrauthorization.New(r.v,
		ecrauthorization.WithCredentials(&r.config.Auth),
		ecrauthorization.WithRepository(r.config.Repository),
	)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get auth provider")
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}
	u.Step(terminal.StatusOK, "successfully fetched access info")

	accessInfo, err := r.getAccessInfo(ctx, authProvider)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get access info")
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}
	u.Step(terminal.StatusOK, "successfully fetched access info")

	if err := r.pushArtifact(ctx, store, accessInfo); err != nil {
		u.Step(terminal.StatusError, "unable to push artifact")
		return nil, fmt.Errorf("unable to push artifact: %w", err)
	}
	u.Step(terminal.StatusOK, "successfully pushed artifact")

	return &terraformv1.Artifact{
		Image:  r.config.Repository,
		Tag:    r.config.Tag,
		Labels: bld.Labels,
	}, nil
}
