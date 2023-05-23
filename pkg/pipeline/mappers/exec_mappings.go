package mappers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

// execInitFn is a function that just does an init, and does not return output
type execInitFn func(context.Context) error

func (p execInitFn) exec(ctx context.Context, l *log.Logger, ui terminal.UI) ([]byte, error) {
	err := p(ctx)
	return nil, err
}

// execPlanFn is a function that returns a terraform plan as a response
type execPlanFn func(context.Context) (*tfjson.Plan, error)

func (p execPlanFn) exec(ctx context.Context, l *log.Logger, ui terminal.UI) ([]byte, error) {
	plan, err := p(ctx)
	if err != nil {
		return nil, err
	}

	byts, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return byts, nil
}

// execOutputFn is a function that returns terraform outputs as a response
type execOutputFn func(context.Context) (map[string]tfexec.OutputMeta, error)

func (p execOutputFn) exec(ctx context.Context, l *log.Logger, ui terminal.UI) ([]byte, error) {
	output, err := p(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	byts, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return byts, nil
}

// execStateFn is a function that returns terraform state as response
type execStateFn func(context.Context) (*tfjson.State, error)

func (p execStateFn) exec(ctx context.Context, l *log.Logger, ui terminal.UI) ([]byte, error) {
	state, err := p(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get state: %w", err)
	}

	byts, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal state: %w", err)
	}

	return byts, nil
}
