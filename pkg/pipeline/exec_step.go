package pipeline

import (
	"context"
	"fmt"
)

func (p *Pipeline) execStep(ctx context.Context, step *Step) error {
	execFn, err := p.mapper.GetExecFn(ctx, step.ExecFn)
	if err != nil {
		return fmt.Errorf("unable to get exec fn: %w", err)
	}

	callbackFn, err := p.mapper.GetCallbackFn(ctx, step.CallbackFn)
	if err != nil {
		return fmt.Errorf("unable to get callback fn: %w", err)
	}

	byts, err := execFn(ctx, p.log, p.ui)
	if err != nil {
		return fmt.Errorf("unable to execute: %w", err)
	}

	if err := callbackFn(ctx, p.log, p.ui, byts); err != nil {
		return fmt.Errorf("unable to execute callback: %w", err)
	}

	return nil
}
