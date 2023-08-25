package run

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/pipeline"
	callbackmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/callbacks"
	execmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/exec"
)

func (r *run) Destroy(ctx context.Context) error {
	pipe, err := r.getDestroyPipeline()
	if err != nil {
		return fmt.Errorf("unable to get destroy pipeline: %w", err)
	}

	if err := pipe.Run(ctx); err != nil {
		return fmt.Errorf("unable to execute destroy pipeline: %w", err)
	}

	return nil
}

func (r *run) getDestroyPipeline() (*pipeline.Pipeline, error) {
	// initialize steps to load the workspace
	pipe, err := pipeline.New(r.v,
		pipeline.WithLogger(r.Log),
		pipeline.WithUI(r.UI),
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
		Name:       "init",
		ExecFn:     execmappers.MapInitLog(r.Workspace.Init),
		CallbackFn: callbackmappers.Noop,
	})

	destroyCb, err := r.outputCallback("destroy.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create output callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "destroy",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.Destroy),
		CallbackFn: destroyCb,
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

	stateCb, err := r.outputCallback("state.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create state callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get state",
		ExecFn:     execmappers.MapTerraformState(r.Workspace.Show),
		CallbackFn: stateCb,
	})
	return pipe, nil
}
