package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	hlog := log.NewHClog(l)

	wkspace, err := p.getWorkspace()
	if err != nil {
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.state.tfWorkspace = wkspace

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(hlog),
		run.WithOutputSettings(&run.OutputSettings{
			Credentials:    &p.state.cfg.Outputs.Auth,
			Bucket:         p.state.cfg.Outputs.Bucket,
			JobPrefix:      p.state.cfg.Outputs.JobPrefix,
			InstancePrefix: p.state.cfg.Outputs.InstancePrefix,
		}),
	)
	if err != nil {
		p.writeErrorResult(ctx, "create terraform run", err)
		return fmt.Errorf("unable to create run: %w", err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeApply:
		l.Info("executing terraform apply")
		err = tfRun.Apply(ctx)
	case models.AppRunnerJobOperationTypeDestroy:
		l.Info("executing terraform destroy")
		err = tfRun.Destroy(ctx)
	case models.AppRunnerJobOperationTypePlanDashOnly:
		l.Info("executing terraform plan")
		err = tfRun.Plan(ctx)
	default:
		l.Error("unsupported terraform run type", zap.String("type", string(p.state.cfg.RunType)))
		return fmt.Errorf("unsupported run type %s", p.state.cfg.RunType)
	}
	if err != nil {
		l.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", p.state.cfg.RunType, err)
	}

	return nil
}
