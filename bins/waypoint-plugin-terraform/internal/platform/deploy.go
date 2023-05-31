package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) deploy(
	ctx context.Context,
	ji *component.JobInfo,
	// artifact is the output of the push function, for the registry
	artifact *terraformv1.Artifact,
	ui terminal.UI,
) (*terraformv1.Deployment, error) {
	tfRun, err := run.New(p.v, run.WithWorkspace(p.Workspace))
	if err != nil {
		return nil, fmt.Errorf("unable to create run: %w", err)
	}

	if err := tfRun.Apply(ctx); err != nil {
		return nil, fmt.Errorf("unable to apply terraform: %w", err)
	}

	return nil, nil
}
