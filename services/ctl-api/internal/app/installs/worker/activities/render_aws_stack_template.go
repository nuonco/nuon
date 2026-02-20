package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/pkg/errors"
)

type RenderAWSStackTemplateRequest struct {
	Input stacks.TemplateInput `temporaljson:"input"`
}

type RenderAWSStackTemplateResponse struct {
	RAWJson  []byte `temporaljson:"raw_json"`
	Checksum string `temporaljson:"checksum"`
}

// @temporal-gen activity
func (a *Activities) RenderAWSStackTemplate(ctx context.Context, req *RenderAWSStackTemplateRequest) (RenderAWSStackTemplateResponse, error) {
	res := RenderAWSStackTemplateResponse{}
	tmpl, awsChecksum, err := a.cfTemplates.Template(&req.Input)
	if err != nil {
		return res, errors.Wrap(err, "unable to create cloudformation template")
	}

	tmplByts, err := tmpl.JSON()
	if err != nil {
		return res, errors.Wrap(err, "unable to get cloudformation json")
	}
	return RenderAWSStackTemplateResponse{
		Checksum: awsChecksum,
		RAWJson:  tmplByts,
	}, nil
}
