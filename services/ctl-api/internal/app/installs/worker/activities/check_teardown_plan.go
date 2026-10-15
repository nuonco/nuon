package activities

import (
	"context"
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/terraform/teardown"
)

type CheckTeardownPlanRequest struct {
	JobID  string   `validate:"required"`
	Retain []string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) CheckTeardownPlan(ctx context.Context, req *CheckTeardownPlanRequest) error {
	job, err := a.GetJob(ctx, &GetJobRequest{ID: req.JobID})
	if err != nil {
		return err
	}
	if job.Execution == nil || job.Execution.Result == nil {
		return temporal.NewNonRetryableApplicationError("teardown plan job has no result", "NoPlan", nil)
	}

	contents, err := job.Execution.Result.GetContentsDisplayDecompressedBytes()
	if err != nil {
		return errors.Wrap(err, "unable to read teardown plan")
	}

	if err := checkTeardownPlanContents(contents, req.Retain); err != nil {
		return temporal.NewNonRetryableApplicationError(err.Error(), "TeardownPlanRejected", err)
	}
	return nil
}

func checkTeardownPlanContents(contents []byte, retain []string) error {
	var plan tfjson.Plan
	if err := json.Unmarshal(contents, &plan); err != nil {
		return errors.Wrap(err, "unable to parse teardown plan to verify retained resources")
	}

	if err := teardown.CheckPlan(&plan, retain); err != nil {
		return errors.Wrap(err, "teardown plan rejected (if this runner predates teardown.retain support, update it and retry)")
	}
	return nil
}
