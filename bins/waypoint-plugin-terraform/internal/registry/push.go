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
	// fetch the source files
	u := ui.Status()
	u.Step(terminal.StatusOK, "fetching source files")
	srcFiles, err := r.getSourceFiles(ctx, src.Path)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get source files")
		return nil, fmt.Errorf("unable to get source files: %w", err)
	}
	u.Step(terminal.StatusOK, "fetched source files")
	u.Close()

	// pack the source files
	u = ui.Status()
	u.Step(terminal.StatusOK, fmt.Sprintf("packing %d files", len(srcFiles)))
	if err := r.packDirectory(ctx, log, u, srcFiles); err != nil {
		u.Step(terminal.StatusError, "unable to pack files")
		return nil, fmt.Errorf("unable to pack files: %w", err)
	}
	u.Step(terminal.StatusOK, "packed store")
	u.Close()

	u = ui.Status()
	authProvider, err := ecrauthorization.New(r.v,
		ecrauthorization.WithCredentials(r.config.Auth),
		ecrauthorization.WithRepository(r.config.Repository),
	)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get auth provider")
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	accessInfo, err := r.getAccessInfo(ctx, authProvider)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get access info")
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}
	u.Step(terminal.StatusOK, fmt.Sprintf("successfully fetched access info %v", accessInfo))

	if err := r.pushArtifact(ctx, accessInfo); err != nil {
		u.Step(terminal.StatusError, "unable to push artifact")
		return nil, fmt.Errorf("unable to push artifact: %w", err)
	}
	u.Step(terminal.StatusOK, "successfully pushed artifact")
	u.Close()

	return &terraformv1.Artifact{
		Image:  r.config.Repository,
		Tag:    r.config.Tag,
		Labels: bld.Labels,
	}, nil
}
