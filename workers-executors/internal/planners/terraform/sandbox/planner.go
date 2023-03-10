package sandbox

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners"
	"go.temporal.io/sdk/log"
)

type planner struct {
	Request *planactivitiesv1.CreateSandboxPlan `validate:"required"`
	L       log.Logger                          `validate:"required"`

	// internal state
	v       *validator.Validate
	sandbox *planv1.Sandbox
}

var _ planners.Planner = (*planner)(nil)

type plannerOption func(*planner) error

func New(v *validator.Validate, opts ...plannerOption) (*planner, error) {
	p := &planner{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating planner: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	if err := p.v.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}

func WithPlan(plan *planactivitiesv1.CreateSandboxPlan) plannerOption {
	return func(p *planner) error {
		p.Request = plan
		p.sandbox = plan.Sandbox
		return nil
	}
}

func WithLogger(l log.Logger) plannerOption {
	return func(p *planner) error {
		p.L = l
		return nil
	}
}
