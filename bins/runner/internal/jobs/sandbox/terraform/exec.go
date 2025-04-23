package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
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

	if err := p.writePolicies(ctx); err != nil {
		return errors.Wrap(err, "unable to write policies")
	}

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(hlog),
		run.WithOutputSettings(&run.OutputSettings{
			Ignore: true,
		}),
	)
	if err != nil {
		p.writeErrorResult(ctx, "create terraform run", err)
		return fmt.Errorf("unable to create run: %w", err)
	}

	if p.state.plan.AWSAuth != nil {
		l.Info("executing with AWS auth " + p.state.plan.AWSAuth.String())
	}
	if p.state.plan.AzureAuth != nil {
		l.Info("executing with Azure auth " + p.state.plan.AzureAuth.String())
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
		l.Error("unsupported terraform run type", zap.String("type", string(job.Operation)))
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}
	if err != nil {
		l.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", job.Operation, err)
	}

	return nil
}
