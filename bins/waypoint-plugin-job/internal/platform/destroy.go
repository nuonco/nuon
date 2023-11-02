package platform

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	jobv1 "github.com/powertoolsdev/mono/pkg/types/plugins/job/v1"
)

func (p *Platform) DestroyFunc() interface{} {
	return p.destroy
}

func (p *Platform) destroy(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
) (*jobv1.Deployment, error) {
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, fmt.Errorf("unable to get output writers")
	}

	_ = hclog.New(&hclog.LoggerOptions{
		Name:   "job",
		Output: stdout,
	})

	// TODO(ja): deployment logic goes here
	return nil, errors.New("not implemented")
}
