package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

func (b *Builder) BuildODRFunc() interface{} {
	return b.buildODR
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger) (*terraformv1.BuildOutput, error) {
	u := ui.Status()
	defer u.Close()

	log.Info("initialized build - log")
	u.Update("initialized build")
	srcFiles, err := b.getSourceFiles(ctx, src.Path)
	if err != nil {
		u.Step(terminal.StatusError, "unable to get source files")
		return nil, fmt.Errorf("unable to get source files: %w", err)
	}
	u.Step(terminal.StatusOK, "fetched source files")

	store, err := b.getStore()
	if err != nil {
		u.Step(terminal.StatusError, "unable to get store")
		return nil, fmt.Errorf("unable to get store: %w", err)
	}
	u.Step(terminal.StatusOK, "created store")

	if err := b.packDirectory(ctx, store, srcFiles); err != nil {
		u.Step(terminal.StatusError, "unable to pack files")
		return nil, fmt.Errorf("unable to pack files: %w", err)
	}
	u.Step(terminal.StatusOK, "packed store")

	if err := store.Close(); err != nil {
		return nil, fmt.Errorf("unable to close store: %w", err)
	}

	return &terraformv1.BuildOutput{
		Files:  srcFiles,
		Labels: b.config.Labels,
	}, nil
}

func (b *Builder) buildODR(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger,
	ai *terraformv1.AccessInfo,
) (*terraformv1.BuildOutput, error) {
	return b.build(ctx, ui, src, log)
}
