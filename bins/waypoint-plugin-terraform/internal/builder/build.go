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

	// push build to local docker
	u.Step(terminal.StatusOK, "noop build")
	return &terraformv1.BuildOutput{
		//Files:	srcFiles,
		Labels: b.config.Labels,
	}, nil

	// TODO(jm): figure out how to reuse the oras cli
	log.Info("packing %d files", len(srcFiles))
	for _, fp := range srcFiles {
		u.Step(terminal.StatusOK, fmt.Sprintf("packing source file %s", fp))
	}
	u.Step(terminal.StatusOK, fmt.Sprintf("packing %d files", len(srcFiles)))
	if err := b.packDirectory(ctx, log, srcFiles); err != nil {
		u.Step(terminal.StatusError, "unable to pack files")
		return nil, fmt.Errorf("unable to pack files: %w", err)
	}
	u.Step(terminal.StatusOK, "packed store")

	// push build to local docker
	return &terraformv1.BuildOutput{
		//Files:	srcFiles,
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
