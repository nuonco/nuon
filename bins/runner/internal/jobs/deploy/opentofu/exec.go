package opentofu

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/kube/config"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/log"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/terraform/run"
	terraformworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Tag this handler's logger with semantic-convention attributes so every
	// emitted record (including from terraform-run helpers below) carries them.
	tfWorkspaceID := ""
	if p.state.plan != nil && p.state.plan.OpentofuDeployPlan != nil && p.state.plan.OpentofuDeployPlan.OpentofuBackend != nil {
		tfWorkspaceID = p.state.plan.OpentofuDeployPlan.OpentofuBackend.WorkspaceID
	}
	l = l.With(
		zap.String("service.name", "runner.opentofu"),
		zap.String("nuon.tool", "opentofu"),
		zap.String("nuon.deploy.kind", "opentofu"),
		zap.String("tf.workspace_id", tfWorkspaceID),
		zap.String("tf.operation", string(job.Operation)),
	)
	ctx = pkgctx.SetLogger(ctx, l)
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
	var wkspace terraformworkspace.Workspace
	if len(planBytes) > 0 {
		l.Info("the plan has ApplyPlanContents, intializing workspace with plan", zap.Int("plan.bytes.count", len(planBytes)))
		wkspace, err = p.GetWorkspaceWithPlan(ctx, planBytes)
		l.Debug("create workspace with plan bytes", zap.Int("plan.bytes.count", len(planBytes)))
	} else {
		l.Info("the plan has no ApplyPlanContents, intializing workspace without plan", zap.Int("plan.bytes.count", len(planBytes)))
		wkspace, err = p.GetWorkspace(ctx)
	}
	if err != nil {
		p.writeErrorResult(ctx, "load opentofu workspace", err)
		return fmt.Errorf("unable to create workspace from config: %w", err)
	}
	p.state.tfWorkspace = wkspace

	// Set the cluster info
	if p.state.plan.OpentofuDeployPlan.ClusterInfo != nil {
		// NOTE(jm): we initialize the root here, because we need to write some state to the directory _before_ we do
		// the run. Ideally this would be handled as part of the lifecycle of the workspace, but it is not yet.
		if err := wkspace.InitRoot(ctx); err != nil {
			return errors.Wrap(err, "unable to initialize root")
		}

		path := filepath.Join(p.state.tfWorkspace.Root(), config.DefaultKubeConfigFilename)
		if err := config.WriteConfig(ctx, p.state.plan.OpentofuDeployPlan.ClusterInfo, path); err != nil {
			return errors.Wrap(err, "unable to write kube config")
		}
	}

	var teardownRetain []string
	if p.state.plan.Teardown != nil {
		teardownRetain = p.state.plan.Teardown.Retain
		l.Info("teardown will retain resources", zap.Strings("tf.teardown.retain", teardownRetain))
	}

	tfRun, err := run.New(p.v, run.WithWorkspace(wkspace),
		run.WithLogger(hclog),
		run.WithOutputSettings(&run.OutputSettings{
			Ignore: true,
		}),
		run.WithTeardownRetain(teardownRetain),
	)
	if err != nil {
		p.writeErrorResult(ctx, "create opentofu run", err)
		return fmt.Errorf("unable to create run: %w", err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		opCtx, end := op.Tool(ctx, "opentofu", "plan")
		opLog := pkgctx.LoggerOrDefault(opCtx, l)
		opLog.Info("executing create opentofu apply plan")
		err = tfRun.Plan(opCtx)
		end(err)
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		opCtx, end := op.Tool(ctx, "opentofu", "destroy_plan")
		opLog := pkgctx.LoggerOrDefault(opCtx, l)
		opLog.Info("executing create opentofu destroy plan")
		err = tfRun.DestroyPlan(opCtx)
		end(err)
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		opCtx, end := op.Tool(ctx, "opentofu", "apply_plan")
		opLog := pkgctx.LoggerOrDefault(opCtx, l)
		opLog.Info("executing opentofu apply")
		err = tfRun.ApplyPlan(opCtx)
		end(err)
	default:
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	if err != nil {
		l.Error("opentofu run errored", zap.Error(err))
		// Persist the full, untruncated tofu error onto the job execution
		// result so ctl-api can parse it into a structured composite error
		// (e.g. a missing AWS IAM permission). The per-execution status
		// description is capped, so this is the authoritative error message.
		p.writeErrorResult(ctx, fmt.Sprintf("opentofu %s", job.Operation), err)
		return fmt.Errorf("unable to execute %s run: %w", job.Operation, err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		if err := p.createResult(ctx, wkspace, hclog); err != nil {
			p.writeErrorResult(ctx, "failed to create sandbox-run install plan", err)
			return err
		}
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		if err := p.createResult(ctx, wkspace, hclog); err != nil {
			p.writeErrorResult(ctx, "failed to create sandbox-run install plan", err)
			return err
		}
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		if err := p.updateTerraformState(ctx, wkspace, hclog); err != nil {
			p.writeErrorResult(ctx, "opentofu show", err)
			// skip returning an error here as the opentofu operation finished successfully & we don't want to fail the job
		}
	}

	return nil
}

func (p *handler) updateTerraformState(ctx context.Context, wkspace terraformworkspace.Workspace, hlog hclog.Logger) error {
	state, err := wkspace.Show(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "opentofu show", err)
		return fmt.Errorf("unable to show state: %w", err)
	}

	// Hand the component's managed cloud resources to the component-health
	// engine: this is the only place the state is readable, since a workspace
	// needs the artifact, backend token and cloud credentials of a deploy job.
	// OpenTofu writes the same state JSON shape as terraform, so the shared
	// terraform provider handles both engines.
	if p.terraformProvider != nil {
		p.terraformProvider.Set(p.state.plan.ComponentID, state)
	}

	stateBody, err := json.Marshal(state)
	if err != nil {
		p.writeErrorResult(ctx, "opentofu show", err)
		return fmt.Errorf("unable to marshal state: %w", err)
	}
	if _, err := p.apiClient.UpdateTerraformStateJSON(
		ctx,
		p.state.plan.OpentofuDeployPlan.OpentofuBackend.WorkspaceID,
		&p.state.jobID,
		stateBody,
	); err != nil {
		p.writeErrorResult(ctx, "opentofu show", err)
		return fmt.Errorf("unable to update state: %w", err)
	}

	return nil
}

// NOTE: createResult is only called in cases when there _is_ a plan. otherwise, we don't really need a result object.
// as a result, we're handling the loading of the plan.json within createResult
func (p *handler) createResult(ctx context.Context, wkspace terraformworkspace.Workspace, hlog hclog.Logger) error {
	// NOTE(fd): the tfplan is already a gzip directory so we do not want to gzip it again.
	// read the tfplan into b64 bytes.
	planBytes, err := wkspace.GetTfplan(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "failed to read tfplan file", err)
		return fmt.Errorf("unable to read tfplan file: %w", err)
	}
	hlog.Info("tfplan", zap.Int("bytes", len(planBytes)))
	planBytesB64 := base64.URLEncoding.EncodeToString(planBytes)
	planJSONBytes, err := wkspace.GetTfplanJsonCompressed(ctx, hlog)
	if err != nil {
		p.writeErrorResult(ctx, "failed to get compressed plan.json bytes", err)
		return fmt.Errorf("unable to read plan.json.gz file: %w", err)
	}
	hlog.Info("plan json", zap.Int("bytes", len(planJSONBytes)))
	planJSONBytesB64 := base64.URLEncoding.EncodeToString(planJSONBytes)
	// create the result object
	_, err = p.apiClient.CreateJobExecutionResult(ctx, p.state.jobID, p.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:                   true,
		ContentsCompressed:        planBytesB64,
		ContentsDisplayCompressed: planJSONBytesB64,
	})
	if err != nil {
		return fmt.Errorf("unable to create opentofu apply job execution result : %w", err)
	}

	return nil
}
