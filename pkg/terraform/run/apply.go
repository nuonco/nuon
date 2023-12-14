package run

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/pipeline"
	callbackmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/callbacks"
	execmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/exec"
)

func (r *run) Apply(ctx context.Context) error {
	pipe, err := r.getApplyPipeline()
	if err != nil {
		return fmt.Errorf("unable to get apply pipeline: %w", err)
	}

	if err := pipe.Run(ctx); err != nil {
		return fmt.Errorf("unable to execute apply pipeline: %w", err)
	}

	return nil
}

func (r *run) instanceOutputCallback(filename string) (pipeline.CallbackFn, error) {
	if r.OutputSettings.Ignore {
		return callbackmappers.Noop, nil
	}
	if r.OutputSettings.InstancePrefix == "" {
		return callbackmappers.Noop, nil
	}

	applyCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.InstancePrefix,
			Filename:     filename,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create apply cb: %w", err)
	}

	return applyCb, nil
}

func (r *run) outputCallback(filename string) (pipeline.CallbackFn, error) {
	if r.OutputSettings.Ignore {
		return callbackmappers.Noop, nil
	}

	applyCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.JobPrefix,
			Filename:     filename,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create apply cb: %w", err)
	}

	return applyCb, nil
}

func (r *run) getApplyPipeline() (*pipeline.Pipeline, error) {
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
		Name:       "load hooks",
		ExecFn:     execmappers.MapInit(r.Workspace.LoadHooks),
		CallbackFn: callbackmappers.Noop,
	})

	pipe.AddStep(&pipeline.Step{
		Name:       "init",
		ExecFn:     execmappers.MapInitLog(r.Workspace.Init),
		CallbackFn: callbackmappers.Noop,
	})

	planCb, err := r.outputCallback("plan.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create plan callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "plan",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.Plan),
		CallbackFn: planCb,
	})

	applyCb, err := r.outputCallback("apply.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create apply callback: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "apply",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.Apply),
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
