package execute

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
)

type ExecutePlanRequest struct {
	Plan *planv1.PlanRef
}

type ExecutePlanResponse struct{}

func (u ExecutePlanRequest) validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func (a *Activities) ExecutePlanAct(ctx context.Context, req ExecutePlanRequest) (ExecutePlanResponse, error) {
	var resp ExecutePlanResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	return resp, nil
}
