package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

func (b *Builder) BuildODRFunc() interface{} {
	return b.buildODR
}

func (b *Builder) buildODR(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	sysLog hclog.Logger,
	accessInfo *ociv1.AccessInfo,
) (*ociv1.BuildOutput, error) {
	// create a logger with the output of the ui
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, fmt.Errorf("unable to get output writers: %w", err)
	}

	log := hclog.New(&hclog.LoggerOptions{
		Name:   "waypoint-plugin-oci",
		Output: stdout,
	})

	log.Info("starting odr build for chart", "path", src.Path)
	b.chartDir = src.Path

	log.Info("packaging chart")
	packagePath, err := b.packageChart(log)
	if err != nil {
		return nil, fmt.Errorf("unable to get source files: %w", err)
	}
	log.Info("successfully packaged chart", "path", packagePath)

	log.Info("creating archive with packaged chart")
	if err := b.packArchive(ctx, log, []fileRef{
		{
			absPath: packagePath,
			relPath: "chart.tgz",
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to pack archive with helm archive: %w", err)
	}
	log.Info("successfully packed archive", "repo", accessInfo.Image)

	log.Info("pushing archive with packaged chart")
	if err := b.pushArchive(ctx, accessInfo); err != nil {
		return nil, fmt.Errorf("unable to pack archive with helm archive: %w", err)
	}
	log.Info("successfully packed archive", "repo", accessInfo.Image)

	log.Info("pushing archive", "repo", accessInfo.Image)
	log.Info("successfully pushed chart", "path", packagePath)
	return &ociv1.BuildOutput{}, nil
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger) (*ociv1.BuildOutput, error) {
	// NOTE: we only use an ODR build because we need the `accessInfo` to be injected
	return nil, fmt.Errorf("only ODR builds supported by this plugin")
}
