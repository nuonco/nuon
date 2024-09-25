package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/terraform/run"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	wkspace, err := p.getWorkspace()
	if err != nil {
		p.log.Error("a-err")
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(p.hclog),
		run.WithOutputSettings(&run.OutputSettings{
			Credentials:    &p.state.cfg.Outputs.Auth,
			Bucket:         p.state.cfg.Outputs.Bucket,
			JobPrefix:      p.state.cfg.Outputs.JobPrefix,
			InstancePrefix: p.state.cfg.Outputs.InstancePrefix,
		}),
	)
	if err != nil {
		p.log.Error("c-err")
		p.writeErrorResult(ctx, "create terraform run", err)
		return fmt.Errorf("unable to create run: %w", err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeApply:
		p.log.Info("executing terraform apply")
		err = tfRun.Apply(ctx)
	case models.AppRunnerJobOperationTypeDestroy:
		p.log.Info("executing terraform destroy")
		err = tfRun.Destroy(ctx)
	case models.AppRunnerJobOperationTypePlanDashOnly:
		p.log.Info("executing terraform plan")
		err = tfRun.Plan(ctx)
	default:
		p.log.Error("unsupported terraform run type", zap.String("type", string(p.state.cfg.RunType)))
		return fmt.Errorf("unsupported run type %s", p.state.cfg.RunType)
	}
	if err != nil {
		p.log.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", p.state.cfg.RunType, err)
	}

	return nil
}
