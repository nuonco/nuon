package run

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline"
	callbackmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/callbacks"
	execmappers "github.com/powertoolsdev/mono/pkg/pipeline/mappers/exec"
	tfo "github.com/powertoolsdev/mono/pkg/terraform/outputs"
	"google.golang.org/protobuf/proto"
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
		Name:       "init",
		ExecFn:     execmappers.MapInitLog(r.Workspace.Init),
		CallbackFn: callbackmappers.Noop,
	})

	applyCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.JobPrefix,
			Filename:     "apply.json",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create apply cb: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "apply",
		ExecFn:     execmappers.MapBytesLog(r.Workspace.Apply),
		CallbackFn: applyCb,
	})

	outputCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.JobPrefix,
			Filename:     "output.json",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create output cb: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get output",
		ExecFn:     execmappers.MapTerraformOutput(r.Workspace.Output),
		CallbackFn: outputCb,
	})

	nuonOutputCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.InstancePrefix,
			Filename:     "output-nuon-v2.pb",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create nuon output cb: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name: "get nuon format output",
		ExecFn: func(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]byte, error) {
			tfOutput, err := r.Workspace.Output(ctx, l)
			if err != nil {
				return nil, err
			}
			pbs, err := tfo.TFOutputMetaToStructPB(tfOutput)
			if err != nil {
				return nil, err
			}
			return proto.Marshal(pbs)
		},
		CallbackFn: nuonOutputCb,
	})

	stateCb, err := callbackmappers.NewS3Callback(r.v,
		callbackmappers.WithCredentials(r.OutputSettings.Credentials),
		callbackmappers.WithBucketKeySettings(callbackmappers.BucketKeySettings{
			Bucket:       r.OutputSettings.Bucket,
			BucketPrefix: r.OutputSettings.JobPrefix,
			Filename:     "state.json",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create output cb: %w", err)
	}
	pipe.AddStep(&pipeline.Step{
		Name:       "get state",
		ExecFn:     execmappers.MapTerraformState(r.Workspace.Show),
		CallbackFn: stateCb,
	})

	return pipe, nil
}
