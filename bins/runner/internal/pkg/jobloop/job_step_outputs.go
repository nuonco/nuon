package jobloop

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (j *jobLoop) getSandboxModePlan(ctx context.Context, job *models.AppRunnerJob) (*plantypes.MinSandboxMode, error) {
	var plan plantypes.MinSandboxMode

	planJSON, err := j.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get job plan")
	}

	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, errors.Wrap(err, "unable to convert to sandbox plan")
	}

	return &plan, nil
}

func (j *jobLoop) sandboxOutputs(ctx context.Context, job *models.AppRunnerJob) (map[string]any, error) {
	plan, err := j.getSandboxModePlan(ctx, job)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get sandbox mode plan")
	}

	if plan.SandboxMode == nil || !plan.SandboxMode.Enabled {
		return map[string]any{}, nil
	}

	return plan.SandboxMode.Outputs, nil
}

func (j *jobLoop) writeTerraformSandboxMode(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution, plan *plantypes.TerraformSandboxMode) error {
	params := url.Values{
		"job_id":       {job.ID},
		"workspace_id": {plan.WorkspaceID},
		"token":        {j.cfg.RunnerAPIToken},
	}

	// TODO(jm): move this into the runner-go-sdk.
	url, err := url.JoinPath(j.cfg.RunnerAPIURL, "/v1/terraform-backend")
	if err != nil {
		return errors.Wrap(err, "unable to get url")
	}
	url = url + "?" + params.Encode()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(plan.StateJSON))
	if err != nil {
		return errors.Wrap(err, "unable to create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "unable to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// write the terraform state to our second endpoint for parsing the resources
	if _, err := j.apiClient.UpdateTerraformStateJSON(ctx, plan.WorkspaceID, &job.ID, []byte(plan.StateJSON)); err != nil {
		return errors.Errorf("unable to update state json")
	}

	if len(plan.PlanContents) > 0 {
		var planDisplayJson *map[string]interface{}
		err = json.Unmarshal([]byte(plan.PlanDisplayContents), &planDisplayJson)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshal plan display")
		}

		// write an output
		if _, err := j.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			Contents:        plan.PlanContents,
			ContentsDisplay: planDisplayJson,
			Success:         true,
		}); err != nil {
			return errors.Wrap(err, "unable to create job execution results")
		}
	}

	return nil
}

func (j *jobLoop) writeHelmSandboxMode(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution, plan *plantypes.HelmSandboxMode) error {
	if len(plan.PlanContents) > 0 {
		var planDisplayJson *map[string]interface{}
		err := json.Unmarshal([]byte(plan.PlanDisplayContents), &planDisplayJson)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshal plan display")
		}
		// write an output
		j.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			Contents:        plan.PlanContents,
			ContentsDisplay: planDisplayJson,
		})
	}

	return nil
}

func (j *jobLoop) executeOutputsJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// write outputs to the api for the job
	var outputs map[string]interface{}
	if j.isSandbox(job) {
		outputs, err = j.sandboxOutputs(ctx, job)
		if err != nil {
			l.Error("unable to get sandbox outputs", zap.Error(err))
			return errors.Wrap(err, "unable to get sandbox outputs")
		}
	} else {
		outputs, err = handler.Outputs(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get outputs")
		}
	}

	_, err = j.apiClient.CreateJobExecutionOutputs(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: outputs,
	})
	if err != nil {
		return errors.Wrap(err, "unable to write outputs to api")
	}

	// for additional sandbox job outputs, make custom requests to fill in data
	if j.isSandbox(job) {
		plan, err := j.getSandboxModePlan(ctx, job)
		if err != nil {
			return errors.Wrap(err, "unable to get sandbox mode plan")
		}

		if plan.SandboxMode != nil && plan.SandboxMode.Terraform != nil {
			if err := j.writeTerraformSandboxMode(ctx, job, jobExecution, plan.SandboxMode.Terraform); err != nil {
				return errors.Wrap(err, "unable to write sandbox mode terraform")
			}
		}
		if plan.SandboxMode != nil && plan.SandboxMode.Helm != nil {
			if err := j.writeHelmSandboxMode(ctx, job, jobExecution, plan.SandboxMode.Helm); err != nil {
				return errors.Wrap(err, "unable to write sandbox mode helm")
			}
		}
	}

	return nil
}
