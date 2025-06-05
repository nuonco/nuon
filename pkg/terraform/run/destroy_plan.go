package run

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/pipeline"
	callbackmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/callbacks"
	execmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/exec"
)

// create a plan for destroy
func (r *run) DestroyPlan(ctx context.Context) error {
	pipe, err := r.getDestroyPipeline()
	if err != nil {
		return fmt.Errorf("unable to get destroy pipeline: %w", err)
	}

	if err := pipe.Run(ctx); err != nil {
		return fmt.Errorf("unable to execute destroy pipeline: %w", err)
	}

	return nil
}

func (r *run) getDestroyPlanPipeline() (*pipeline.Pipeline, error) {
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

	planCb, err := r.outputCallback("tfplan")
	if err != nil {
		return nil, fmt.Errorf("unable to create output callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "plan destroy",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.ApplyDestroyPlan),
		CallbackFn: planCb,
	})

	planJsonCb, err := r.outputCallback("plan.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create plan callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "show plan",
		ExecFn:     execmappers.MapTerraformPlan(r.Workspace.ShowPlan),
		CallbackFn: planJsonCb,
	})
	return pipe, nil
}
