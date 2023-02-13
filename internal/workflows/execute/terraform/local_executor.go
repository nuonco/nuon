package execute

import (
	"context"
	"fmt"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	"github.com/powertoolsdev/workers-executors/internal/executors/terraform"
)

func (a *Activities) ExecuteTerraformPlanLocally(
	ctx context.Context,
	req *executev1.ExecutePlanRequest,
) (*executev1.ExecutePlanResponse, error) {
	executor, err := terraform.New(a.v, terraform.WithPlan(req.Plan))
	if err != nil {
		return nil, fmt.Errorf("unable to get executor: %w", err)
	}

	_, err = executor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("executor did not succeed: %w", err)
	}
	return &executev1.ExecutePlanResponse{}, nil
}
