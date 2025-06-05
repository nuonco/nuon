package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	hlog := log.NewHClog(l)

	if err := p.writePolicies(ctx); err != nil {
		return errors.Wrap(err, "unable to write policies")
	}

	wkspace, err := p.getWorkspace()
	if err != nil {
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}

	// TODO: when we split this up, load the plan into the workspace like this
	// plan := ""
	// wkspace.WritePlan(ctx, plan)

	// assign workspace
	p.state.tfWorkspace = wkspace

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

	// TODO: update these to actulaly plan and apply
	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		l.Info("creating terraform plan")
		err = tfRun.Plan(ctx)
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Info("creating terraform teardown plan")
		err = tfRun.DestroyPlan(ctx)
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l.Info("executing terraform apply plan")
		err = tfRun.ApplyPlan(ctx)
	default:
		l.Error("unsupported terraform run type", zap.String("type", string(job.Operation)))
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}
	if err != nil {
		l.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", job.Operation, err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		if err := p.updateTerraformState(ctx, wkspace, hlog); err != nil {
			p.writeErrorResult(ctx, "terraform show", err)
		}
	}

	return nil
}

func (p *handler) updateTerraformState(ctx context.Context, wkspace workspace.Workspace, hlog hclog.Logger) error {
	state, err := wkspace.Show(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "terraform show", err)
		return fmt.Errorf("unable to show state: %w", err)
	}

	stateBody, err := json.Marshal(state)
	if err != nil {
		p.writeErrorResult(ctx, "terraform show", err)
		return fmt.Errorf("unable to marshal state: %w", err)
	}

	if _, err := p.apiClient.UpdateTerraformStateJSON(ctx, p.state.plan.TerraformBackend.WorkspaceID, &p.state.jobID, stateBody); err != nil {
		p.writeErrorResult(ctx, "terraform show", err)
		return fmt.Errorf("unable to update state: %w", err)
	}

	return nil
}

func (p *handler) loadPlan(ctx context.Context) error {
	// write the plan from p.state.plan <dot> plan to plan.json in the workspace
	// err := p.state.tfWorkspace.WritePlan(ctx, p.state.plan)
	// return err
	return nil
}

func (p *handler) createResult(ctx context.Context) error {
	pathToPlan := p.state.tfWorkspace.Root() + "/" + "plan.json"

	// Read the plan.json file into a string
	planBytes, err := os.ReadFile(pathToPlan)
	if err != nil {
		p.writeErrorResult(ctx, "failed to read plan.json file", err)
		return fmt.Errorf("unable to read plan.json file: %w", err)
	}

	planJSON := string(planBytes)
	_, err = p.apiClient.CreateJobExecutionResult(ctx, p.state.jobID, p.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:  true,
		Contents: planJSON,
	})
	if err != nil {
		return fmt.Errorf("unable to create terraform apply job execution result : %w", err)
	}
	return nil
}
