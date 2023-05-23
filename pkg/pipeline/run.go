package pipeline

import (
	"context"
	"fmt"
)

// Run runs a pipeline from end to end
func (p *Pipeline) Run(ctx context.Context) error {
	if err := p.v.Struct(p); err != nil {
		return fmt.Errorf("invalid pipeline: %w", err)
	}

	for idx, step := range p.Steps {
		if err := p.execStep(ctx, step); err != nil {
			return fmt.Errorf("unable to execute step %d.%s: %w", idx, step.Name, err)
		}
	}

	return nil
}
