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

	if _, err := executor.Execute(ctx); err != nil {
		return nil, fmt.Errorf("unable to execute plan: %w", err)
	}

	return nil, nil
}
