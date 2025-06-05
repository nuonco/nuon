package run

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/pipeline"
	callbackmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/callbacks"
	execmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/exec"
)

func (r *run) ApplyPlan(ctx context.Context) error {
	pipe, err := r.getApplyPlanPipeline()
	if err != nil {
		return fmt.Errorf("unable to get apply pipeline: %w", err)
	}

	if err := pipe.Run(ctx); err != nil {
		return fmt.Errorf("unable to execute apply pipeline: %w", err)
	}

	return nil
}

func (r *run) getApplyPlanPipeline() (*pipeline.Pipeline, error) {
	// initialize steps to load the workspace
	pipe, err := pipeline.New(r.v,
		pipeline.WithLogger(r.Log),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create pipeline: %w", err)
	}

	pipe.AddStep(&pipeline.Step{
		Name:       "initialize workspace",
		ExecFn:     execmappers.MapInit(r.Workspace.InitRoot),
		CallbackFn: callbackmappers.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load archive",
		ExecFn:     execmappers.MapInit(r.Workspace.LoadArchive),
		CallbackFn: callbackmappers.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load backend",
		ExecFn:     execmappers.MapInit(r.Workspace.LoadBackend),
		CallbackFn: callbackmappers.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load binary",
		ExecFn:     execmappers.MapInitLog(r.Workspace.LoadBinary),
		CallbackFn: callbackmappers.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load variables",
		ExecFn:     execmappers.MapInit(r.Workspace.LoadVariables),
		CallbackFn: callbackmappers.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load hooks",
		ExecFn:     execmappers.MapInit(r.Workspace.LoadHooks),
		CallbackFn: callbackmappers.Noop,
	})

	pipe.AddStep(&pipeline.Step{
		Name:       "init",
		ExecFn:     execmappers.MapInitLog(r.Workspace.Init),
		CallbackFn: callbackmappers.Noop,
	})

	applyCb, err := r.outputCallback("apply.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create apply callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "apply plan",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.ApplyPlan),
		CallbackFn: applyCb,
	})

	outputCb, err := r.outputCallback("output.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create output callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get output",
		ExecFn:     execmappers.MapTerraformOutput(r.Workspace.Output),
		CallbackFn: outputCb,
	})

	structOutputCb, err := r.outputCallback("output-struct-v1.pb")
	if err != nil {
		return nil, fmt.Errorf("unable to create struct output callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get struct output",
		ExecFn:     execmappers.MapStructOutput(r.Workspace.Output),
		CallbackFn: structOutputCb,
	})

	instanceStructOutputCb, err := r.instanceOutputCallback("output-struct-v1.pb")
	if err != nil {
		return nil, fmt.Errorf("unable to create struct output callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get struct output",
		ExecFn:     execmappers.MapStructOutput(r.Workspace.Output),
		CallbackFn: instanceStructOutputCb,
	})

	stateCb, err := r.outputCallback("state.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create state callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get state",
		ExecFn:     execmappers.MapTerraformState(r.Workspace.Show),
		CallbackFn: stateCb,
	})

	instanceStateCb, err := r.instanceOutputCallback("state.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create state callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get state",
		ExecFn:     execmappers.MapTerraformState(r.Workspace.Show),
		CallbackFn: instanceStateCb,
	})

	return pipe, nil
}
