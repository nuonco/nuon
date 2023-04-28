package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

type BuildConfig struct {
	OutputName string `hcl:"output_name,optional"`
	Source     string `hcl:"source,optional"`
}

type Builder struct {
	config BuildConfig
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	c, ok := config.(*BuildConfig)
	if !ok {
		return fmt.Errorf("expected type BuildConfig")
	}

	_, err := os.Stat(c.Source)
	if err != nil {
		return fmt.Errorf("source folder does not exist")
	}

	return nil
}

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context, ui terminal.UI, log hclog.Logger) (*terraformv1.BuildOutput, error) {
	u := ui.Status()
	defer u.Close()

	u.Update("terraform plugin")
	u.Step(terminal.StatusError, "dummy error")
	u.Step(terminal.StatusOK, "Application built successfully")

	return &terraformv1.BuildOutput{}, nil
}
