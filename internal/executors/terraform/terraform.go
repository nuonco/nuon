package terraform

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-terraform/pkg/runner"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

type terraformer struct {
	Plan *planv1.PlanRef `validate:"required:dive"`

	// internal state
	v *validator.Validate
}

type terraformerOption func(*terraformer) error

func New(v *validator.Validate, opts ...terraformerOption) (*terraformer, error) {
	t := &terraformer{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating terraformer: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	if err := t.v.Struct(t); err != nil {
		return nil, err
	}

	return t, nil
}

// WithPlan specifies the location of the terraform plan to execute
func WithPlan(p *planv1.PlanRef) terraformerOption {
	return func(t *terraformer) error {
		t.Plan = p
		return nil
	}
}

func (t *terraformer) Run(ctx context.Context) (map[string]interface{}, error) {
	r, err := runner.New(
		t.v,
		runner.WithPlan(t.Plan),
	)

	if err != nil {
		return nil, err
	}

	return r.Run(ctx)
}
