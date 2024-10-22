package job

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

func (p *handler) DestroyFunc() interface{} {
	return p.destroy
}

func (p *handler) destroy(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
) error {
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return fmt.Errorf("unable to get output writers")
	}

	_ = hclog.New(&hclog.LoggerOptions{
		Name:   "job",
		Output: stdout,
	})

	// TODO(ja): deployment logic goes here
	return errors.New("not implemented")
}
