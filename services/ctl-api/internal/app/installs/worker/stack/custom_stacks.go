package stack

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// RenderAndUploadCustomStacksTemplate renders and uploads the custom-stacks-only
// CloudFormation artifact for the Terraform install path.
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

	customRendered, err := activities.AwaitRenderAWSStackTemplate(ctx, &activities.RenderAWSStackTemplateRequest{
		Input: inp,
	})
	if err != nil {
		return errors.Wrap(err, "unable to render custom stacks only template")
	}

	// BuildInstallerSDKConfig reads it back off stackVersion instead of
	// re-fetching/parsing templates on every terraform plan.
	if err := activities.AwaitSaveInstallStackVersionCustomStacksOutputMap(ctx, &activities.SaveInstallStackVersionCustomStacksOutputMapRequest{
		ID:                 stackVersion.ID,
		OutputMap:          customRendered.CustomStacksOutputMap,
		InputParametersMap: customRendered.CustomStacksInputParametersMap,
	}); err != nil {
		return errors.Wrap(err, "unable to save custom stacks output map")
	}

	if templateBucket == "" {
		workflow.GetLogger(ctx).Warn("cloudformation stack template bucket not configured, skipping custom stacks only s3 upload", "install_id", installID)
		return nil
	}

	if err := activities.AwaitUploadAWSCloudFormationStackVersionTemplate(ctx, &activities.UploadAWSCloudFormationStackVersionTemplateRequest{
		BucketKey: stackVersion.CustomStacksAWSBucketKey,
		Template:  customRendered.RAWJson,
	}); err != nil {
		return errors.Wrap(err, "unable to upload custom stacks only template")
	}

	return nil
}
