package terraform

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/pkg/kube/config"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}
	hclog := log.NewHClog(l)

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
		wkspace, err = p.GetWorkspaceWithPlan(ctx, planBytes)
		l.Debug("create workspace with plan bytes", zap.Int("plan.bytes.count", len(planBytes)))
	} else {
		l.Info("the plan has no ApplyPlanContents, intializing workspace without plan", zap.Int("plan.bytes.count", len(planBytes)))
		wkspace, err = p.GetWorkspace(ctx)
	}
	if err != nil {
		p.writeErrorResult(ctx, "load terraform workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.state.tfWorkspace = wkspace

	// Set the cluster info
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
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		l.Info("executing create terraform apply plan")
		err = tfRun.Plan(ctx)
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Info("executing create terraform destroy plan")
		err = tfRun.DestroyPlan(ctx)
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l.Info("executing terraform apply")
		err = tfRun.ApplyPlan(ctx)
	default:
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	if err != nil {
		l.Error("terraform run errored", zap.Error(err))
		return fmt.Errorf("unable to execute %s run: %w", job.Operation, err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		if err := p.createResult(ctx); err != nil {
			p.writeErrorResult(ctx, "failed to create sandbox-run install plan", err)
			return err
		}
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		if err := p.createResult(ctx); err != nil {
			p.writeErrorResult(ctx, "failed to create sandbox-run install plan", err)
			return err
		}
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		if err := p.updateTerraformState(ctx, wkspace, hclog); err != nil {
			p.writeErrorResult(ctx, "terraform show", err)
			// skip returning an error here as the terraform operation finished successfully & we don't want to fail the job
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

	if _, err := p.apiClient.UpdateTerraformStateJSON(ctx, p.state.plan.TerraformDeployPlan.TerraformBackend.WorkspaceID, &p.state.jobID, stateBody); err != nil {
		p.writeErrorResult(ctx, "terraform show", err)
		return fmt.Errorf("unable to update state: %w", err)
	}

	return nil
}

// NOTE: createResult is only called in cases when there _is_ a plan. otherwise, we don't really need a result object.
// as a result, we're handling the loading of the plan.json within createResult
func (p *handler) createResult(ctx context.Context) error {
	// read the tfplan into b64 byts
	pathToPlan := filepath.Join(p.state.tfWorkspace.Root(), "tfplan") // TODO: make this a built in on the workspace (tfplan)
	planBytes, err := os.ReadFile(pathToPlan)
	if err != nil {
		p.writeErrorResult(ctx, "failed to read tfplan file", err)
		return fmt.Errorf("unable to read tfplan file: %w", err)
	}
	planContents := base64.StdEncoding.EncodeToString(planBytes)

	// read the plan.json into Display
	pathToPlanJson := filepath.Join(p.state.tfWorkspace.Root(), "plan.json") // TODO: make this a built in on the workspace (GetPlan)
	planJsonString, err := os.ReadFile(pathToPlanJson)
	var planJson *map[string]interface{}
	err = json.Unmarshal(planJsonString, &planJson)

	if err != nil {
		p.writeErrorResult(ctx, "failed to read plan.json file", err)
		return fmt.Errorf("unable to read plan.json file: %w", err)
	}
	// create the result object
	_, err = p.apiClient.CreateJobExecutionResult(ctx, p.state.jobID, p.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:         true,
		Contents:        planContents,
		ContentsDisplay: planJson,
	})
	if err != nil {
		return fmt.Errorf("unable to create terraform apply job execution result : %w", err)
	}

	return nil
}
