package execute

import (
	"context"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
)

func (a *Activities) ExecutePlanLocally(ctx context.Context, req *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	return nil, nil
}
