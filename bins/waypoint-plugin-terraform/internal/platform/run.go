package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

type runType int

const (
	runTypeApply runType = iota + 1
	runTypePlan
	runTypeDestroy
)

func (r runType) String() string {
	switch r {
	case runTypeApply:
		return "apply"
	case runTypePlan:
		return "plan"
	case runTypeDestroy:
		return "destroy"
	}

	return ""
}

func (p *Platform) execRun(
	ctx context.Context,
	runTyp runType,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
) (*terraformv1.Deployment, error) {
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, fmt.Errorf("unable to get output writers: %w", err)
	}

	runLog := hclog.New(&hclog.LoggerOptions{
		Name:   "terraform",
		Output: stdout,
	})

	tfRun, err := run.New(p.v, run.WithWorkspace(p.Workspace),
		run.WithUI(ui),
		run.WithLogger(runLog),
		run.WithOutputSettings(&run.OutputSettings{
			Credentials:    &p.Cfg.Outputs.Auth,
			Bucket:         p.Cfg.Outputs.Bucket,
			JobPrefix:      p.Cfg.Outputs.JobPrefix,
			InstancePrefix: p.Cfg.Outputs.InstancePrefix,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create run: %w", err)
	}

	switch runTyp {
	case runTypeApply:
		err = tfRun.Apply(ctx)
	case runTypeDestroy:
		err = tfRun.Destroy(ctx)
	case runTypePlan:
		err = tfRun.Plan(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to execute %s run: %w", runTyp, err)
	}

	return &terraformv1.Deployment{
		StateKey:    p.Cfg.Backend.StateKey,
		StateBucket: p.Cfg.Backend.Bucket,
		Labels:      p.Cfg.Labels,
	}, nil
}
