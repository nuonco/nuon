package run

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/pipeline"
	"github.com/powertoolsdev/mono/pkg/pipeline/callbacks"
)

// Plan will initialize the workspace and then execute functions in it
func (r *run) Plan(ctx context.Context) error {
	pipe, err := r.getPlanPipeline()
	if err != nil {
		return fmt.Errorf("unable to get plan pipeline: %w", err)
	}

	if err := pipe.Run(ctx); err != nil {
		return fmt.Errorf("unable execute plan pipeline: %w", err)
	}

	return nil
}

func (r *run) getPlanPipeline() (*pipeline.Pipeline, error) {
	pipe, err := pipeline.New(r.v)
	if err != nil {
		return nil, fmt.Errorf("unable to get destroy pipeline: %w", err)
	}

	pipe.AddStep(&pipeline.Step{
		Name:       "initialize workspace",
		ExecFn:     r.Workspace.Init,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load archive",
		ExecFn:     r.Workspace.LoadArchive,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load backend",
		ExecFn:     r.Workspace.LoadBackend,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load binary",
		ExecFn:     r.Workspace.LoadBinary,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "load variables",
		ExecFn:     r.Workspace.LoadVariables,
		CallbackFn: callbacks.Noop,
	})

	pipe.AddStep(&pipeline.Step{
		Name:       "plan",
		ExecFn:     r.Workspace.Plan,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "get output",
		ExecFn:     r.Workspace.Output,
		CallbackFn: callbacks.Noop,
	})
	pipe.AddStep(&pipeline.Step{
		Name:       "get state",
		ExecFn:     r.Workspace.Show,
		CallbackFn: callbacks.Noop,
	})

	return pipe, nil
}
