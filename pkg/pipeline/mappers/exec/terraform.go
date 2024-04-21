package exec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline"
	tfo "github.com/powertoolsdev/mono/pkg/terraform/outputs"
	"google.golang.org/protobuf/proto"
)

func MapTerraformPlan(fn execPlanFn) pipeline.ExecFn {
	return fn.exec
}

// execPlanFn is a function that returns a terraform plan as a response
type execPlanFn func(context.Context, hclog.Logger) (*tfjson.Plan, error)

func (p execPlanFn) exec(ctx context.Context, log hclog.Logger, ui terminal.UI) ([]byte, error) {
	plan, err := p(ctx, log)
	if err != nil {
		return nil, err
	}

	byts, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return byts, nil
}

func MapTerraformOutput(fn execOutputFn) pipeline.ExecFn {
	return fn.exec
}

// execOutputFn is a function that returns terraform outputs as a response
type execOutputFn func(context.Context, hclog.Logger) (map[string]tfexec.OutputMeta, error)

func (p execOutputFn) exec(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]uint8, error) {
	output, err := p(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	byts, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return byts, nil
}

func MapStructOutput(fn execOutputFn) pipeline.ExecFn {
	return fn.mapStruct
}

func (p execOutputFn) mapStruct(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]uint8, error) {
	tfOutput, err := p(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	protoStruct, err := tfo.TFOutputMetaToStructPB(tfOutput)
	if err != nil {
		return nil, fmt.Errorf("unable to convert TF output to structpb: %w", err)
	}

	return proto.Marshal(protoStruct)
}

func MapTerraformState(fn execStateFn) pipeline.ExecFn {
	return fn.exec
}

// execStateFn is a function that returns terraform state as response
type execStateFn func(context.Context, hclog.Logger) (*tfjson.State, error)

func (p execStateFn) exec(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]uint8, error) {
	state, err := p(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("unable to get state: %w", err)
	}

	byts, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal state: %w", err)
	}

	return byts, nil
}

func MapTerraformValidate(fn execValidateFn) pipeline.ExecFn {
	return fn.exec
}

// execValidateFn is a function that returns terraform validation as the response
type execValidateFn func(context.Context, hclog.Logger) (*tfjson.ValidateOutput, error)

func (p execValidateFn) exec(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]uint8, error) {
	out, err := p(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", err)
	}

	byts, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal state: %w", err)
	}

	return byts, nil
}
