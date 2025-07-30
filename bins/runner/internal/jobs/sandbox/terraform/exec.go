package terraform

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

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

	// Load Plan Bytes
	var planBytes []byte
	if len(p.state.plan.ApplyPlanContents) > 0 {
		b64EncodedContent := p.state.plan.ApplyPlanContents
		planBytes, err = base64.StdEncoding.DecodeString(b64EncodedContent)
		if err != nil {
			return errors.Wrap(err, "unable to decode base64 Plan.Contents into bytes.")
		}
	} else {
		planBytes = []byte{}
	}

	// get the right workspace
	var wkspace workspace.Workspace
	if len(planBytes) > 0 {
		l.Info("the plan has ApplyPlanContents, intializing workspace with plan", zap.Int("plan.bytes.count", len(planBytes)))
		wkspace, err = p.getWorkspaceWithPlan(planBytes)
		l.Debug("create workspace with plan bytes", zap.Int("plan.bytes.count", len(planBytes)))
	} else {
		l.Info("the plan has no ApplyPlanContents, intializing workspace without plan", zap.Int("plan.bytes.count", len(planBytes)))
		wkspace, err = p.getWorkspace()
	}

	// initialize
	if err := wkspace.InitRoot(ctx); err != nil {
		return errors.Wrap(err, "unable to initialize root")
	}
	if err != nil {
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	} else {
		l.Info(
			"workspace acquired",
			zap.String("root", wkspace.Root()),
		)
	}

	// assign workspace
	p.state.tfWorkspace = wkspace

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(hlog),
		run.WithOutputSettings(&run.OutputSettings{ // TODO: remove entirely - this is for S3
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
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		l.Info("creating terraform plan", zap.String("operation", string(job.Operation)))
		err = tfRun.Plan(ctx)
		p.createJobExecutionResultRequest(ctx, wkspace, hlog)
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Info("creating terraform teardown plan", zap.String("operation", string(job.Operation)))
		err = tfRun.DestroyPlan(ctx)
		p.createJobExecutionResultRequest(ctx, wkspace, hlog)
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l.Info("executing terraform apply plan", zap.String("operation", string(job.Operation)))
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

// NOTE: createJobExecutionResultRequest is only called in cases when there _is_ a plan. otherwise, we don't really need a result object.
// as a result, we're handling the loading of the plan.json within createJobExecutionResultRequest
func (p *handler) createJobExecutionResultRequest(ctx context.Context, wkspace workspace.Workspace, hlog hclog.Logger) error {
	// NOTE(fd): the tfplan is already a gzip directory so we do not want to gzip it again.
	// read the tfplan into b64 bytes.
	planBytes, err := wkspace.GetTfplan(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "failed to read tfplan file", err)
		return fmt.Errorf("unable to read tfplan file: %w", err)
	}
	hlog.Info("tfplan", zap.Int("bytes", len(planBytes)))
	planBytesB64 := base64.URLEncoding.EncodeToString(planBytes)
	planJsonBytes, err := wkspace.GetTfplanJsonCompressed(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "failed to get compressed plan.json.gz bytes", err)
		return fmt.Errorf("unable to read plan.json.gz file: %w", err)
	}
	hlog.Info("plan json", zap.Int("bytes", len(planJsonBytes)))
	planJsonBytesB64 := base64.URLEncoding.EncodeToString(planJsonBytes)
	// create the result object
	_, err = p.apiClient.CreateJobExecutionResult(ctx, p.state.jobID, p.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:                   true,
		ContentsCompressed:        planBytesB64,
		ContentsDisplayCompressed: planJsonBytesB64,
	})
	if err != nil {
		return fmt.Errorf("unable to create terraform apply job execution result : %w", err)
	}

	// return nil
	return nil
}
