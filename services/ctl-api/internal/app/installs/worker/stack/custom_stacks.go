package stack

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func RenderAndUploadCustomStacksTemplate(
	ctx workflow.Context,
	installID string,
	stackVersion *app.InstallStackVersion,
	inp stacks.TemplateInput,
	templateBucket string,
) error {
	if stackVersion.CustomStacksAWSBucketKey == "" {
		return nil
	}

	inp.CustomStacksOnly = true

	var template []byte
	var outputMap map[string]map[string]string
	var inputParametersMap map[string]map[string]string
	switch inp.AppCfg.RunnerConfig.Type {
	case app.AppRunnerTypeAWS:
		customRendered, err := activities.AwaitRenderAWSStackTemplate(ctx, &activities.RenderAWSStackTemplateRequest{
			Input: inp,
		})
		if err != nil {
			return errors.Wrap(err, "unable to render custom stacks only template")
		}
		template = customRendered.RAWJson
		outputMap = customRendered.CustomStacksOutputMap
		inputParametersMap = customRendered.CustomStacksInputParametersMap
	case app.AppRunnerTypeAzure:
		customRendered, err := activities.AwaitRenderARMStackTemplate(ctx, &activities.RenderARMStackTemplateRequest{
			Input: inp,
		})
		if err != nil {
			return errors.Wrap(err, "unable to render custom stacks only template")
		}
		template = customRendered.RAWJson
		outputMap = customRendered.CustomStacksOutputMap
		inputParametersMap = customRendered.CustomStacksInputParametersMap
	default:
		return nil
	}

	if err := activities.AwaitSaveInstallStackVersionCustomStacksOutputMap(ctx, &activities.SaveInstallStackVersionCustomStacksOutputMapRequest{
		ID:                 stackVersion.ID,
		OutputMap:          outputMap,
		InputParametersMap: inputParametersMap,
	}); err != nil {
		return errors.Wrap(err, "unable to save custom stacks output map")
	}

	if templateBucket == "" {
		workflow.GetLogger(ctx).Warn("cloudformation stack template bucket not configured, skipping custom stacks only s3 upload", "install_id", installID)
		return nil
	}

	if err := activities.AwaitUploadAWSCloudFormationStackVersionTemplate(ctx, &activities.UploadAWSCloudFormationStackVersionTemplateRequest{
		BucketKey: stackVersion.CustomStacksAWSBucketKey,
		Template:  template,
	}); err != nil {
		return errors.Wrap(err, "unable to upload custom stacks only template")
	}

	return nil
}
