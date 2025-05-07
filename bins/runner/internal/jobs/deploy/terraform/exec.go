package terraform

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/pkg/kube/config"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}
	hclog := log.NewHClog(l)

	wkspace, err := p.GetWorkspace(ctx)
	if err != nil {
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.state.tfWorkspace = wkspace

	if p.state.plan.TerraformDeployPlan.ClusterInfo != nil {
		// NOTE(jm): we initialize the root here, because we need to write some state to the directory _before_ we do
		// the run. Ideally this would be handled as part of the lifecycle of the workspace, but it is not yet.
		if err := wkspace.InitRoot(ctx); err != nil {
			return errors.Wrap(err, "unable to initialize root")
		}

		path := filepath.Join(p.state.tfWorkspace.Root(), config.DefaultKubeConfigFilename)
		if err := config.WriteConfig(ctx, p.state.plan.TerraformDeployPlan.ClusterInfo, path); err != nil {
			return errors.Wrap(err, "unable to write kube config")
		}
	}

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(hclog),
		run.WithOutputSettings(&run.OutputSettings{
			Ignore: true,
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
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	if err != nil {
		l.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", job.Operation, err)
	}

	return nil
}
