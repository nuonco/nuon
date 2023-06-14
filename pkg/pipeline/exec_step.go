package pipeline

import (
	"context"
	"fmt"
)

func (p *Pipeline) execStep(ctx context.Context, step *Step) error {
	p.Log.Info("executing step ", "name", step.Name)
	if err := p.v.Struct(step); err != nil {
		return fmt.Errorf("unable to validate step: %w", err)
	}

	byts, err := step.ExecFn(ctx, p.Log, p.UI)
	if err != nil {
		return fmt.Errorf("unable to execute: %w", err)
	}

	if err := step.CallbackFn(ctx, p.Log, p.UI, byts); err != nil {
		return fmt.Errorf("unable to execute callback: %w", err)
	}

	return nil
}
