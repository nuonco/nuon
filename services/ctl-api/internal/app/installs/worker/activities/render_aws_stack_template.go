package activities

import (
	"context"
	"regexp"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

type RenderAWSStackTemplateRequest struct {
	Input stacks.TemplateInput `temporaljson:"input"`
}

type RenderAWSStackTemplateResponse struct {
	RAWJson  []byte `temporaljson:"raw_json"`
	Checksum string `temporaljson:"checksum"`
}

// @temporal-gen-v2 activity
func (a *Activities) RenderAWSStackTemplate(ctx context.Context, req *RenderAWSStackTemplateRequest) (RenderAWSStackTemplateResponse, error) {
	res := RenderAWSStackTemplateResponse{}

	inp := req.Input
	if err := a.setTargetAWSAccount(ctx, &inp); err != nil {
		return res, err
	}

	tmplByts, awsChecksum, err := a.cfTemplates.Template(&inp)
	if err != nil {
		return res, errors.Wrap(err, "unable to create cloudformation template")
	}

	return RenderAWSStackTemplateResponse{
		Checksum: awsChecksum,
		RAWJson:  tmplByts,
	}, nil
}

// setTargetAWSAccount pins the rendered template to the install's target AWS
// account when the org has phone-home-auth enabled. The 12-digit guard is not
// redundant with creation-time validation: the target is only validated when the
// flag is on, so an install created before the flag was flipped can carry an
// arbitrary value — pinning to it would make the stack undeployable everywhere.
func (a *Activities) setTargetAWSAccount(ctx context.Context, inp *stacks.TemplateInput) error {
	target := inp.Install.CloudPlatformMetadata.TargetAccountID
	if target == "" {
		return nil
	}

	enabled, err := a.features.OrgHasFeature(ctx, inp.Install.OrgID, app.OrgFeaturePhoneHomeAuth)
	if err != nil {
		return errors.Wrap(err, "unable to check phone home auth feature")
	}
	if !enabled {
		return nil
	}

	if !awsAccountIDPattern.MatchString(target) {
		a.l.Warn("not pinning stack template to malformed target aws account id",
			zap.String("install_id", inp.Install.ID),
			zap.String("org_id", inp.Install.OrgID),
		)
		return nil
	}

	inp.TargetAWSAccountID = target
	return nil
}

var awsAccountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)
