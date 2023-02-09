package execute

import (
	"context"
	"fmt"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint"
	"go.uber.org/zap"
)

func (a *Activities) ExecutePlanLocally(ctx context.Context, req *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	zapLog := zap.L()

	executor, err := waypoint.New(a.v, waypoint.WithPlan(req.Plan),
		waypoint.WithLogger(zapLog))
	if err != nil {
		return nil, fmt.Errorf("unable to get executor: %w", err)
	}

	resp, err := executor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("executor did not succeed: %w", err)
	}
	return resp, nil
}
