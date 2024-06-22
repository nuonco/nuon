package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (p *Platform) execRun(
	ctx context.Context,
	ji *component.JobInfo,
	src *component.Source,
	ui terminal.UI,
	log hclog.Logger,
) (*terraformv1.Deployment, error) {
	p.Path = src.Path
	wkspace, err := p.GetWorkspace()
	if err != nil {
		return &terraformv1.Deployment{}, fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.Workspace = wkspace

	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return &terraformv1.Deployment{}, fmt.Errorf("unable to get output writers: %w", err)
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
		return &terraformv1.Deployment{}, fmt.Errorf("unable to create run: %w", err)
	}

	switch p.Cfg.RunType {
	case configs.TerraformDeployRunTypeApply:
		runLog.Info("executing terraform apply")
		err = tfRun.Apply(ctx)
	case configs.TerraformDeployRunTypeDestroy:
		runLog.Info("executing terraform destroy")
		err = tfRun.Destroy(ctx)
	case configs.TerraformDeployRunTypePlan:
		runLog.Info("executing terraform plan")
		err = tfRun.Plan(ctx)
	default:
		runLog.Error("unsupported terraform run type %s", p.Cfg.RunType)
		return &terraformv1.Deployment{}, fmt.Errorf("unsupported run type %s", p.Cfg.RunType)
	}
	if err != nil {
		runLog.Error("terraform run errored - %s", err.Error())
		return &terraformv1.Deployment{}, fmt.Errorf("unable to execute %s run: %w", p.Cfg.RunType, err)
	}

	return &terraformv1.Deployment{
		StateKey:    p.Cfg.Backend.StateKey,
		StateBucket: p.Cfg.Backend.Bucket,
		Labels:      p.Cfg.Labels,
	}, nil
}
