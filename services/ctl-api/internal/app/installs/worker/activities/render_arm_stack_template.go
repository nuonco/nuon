package activities

import (
	"context"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/arm"
)

type RenderARMStackTemplateRequest struct {
	Input stacks.TemplateInput `temporaljson:"input"`
}

type RenderARMStackTemplateResponse struct {
	RAWJson  []byte `temporaljson:"raw_json"`
	Checksum string `temporaljson:"checksum"`
}

// @temporal-gen-v2 activity
func (a *Activities) RenderARMStackTemplate(ctx context.Context, req *RenderARMStackTemplateRequest) (RenderARMStackTemplateResponse, error) {
	res := RenderARMStackTemplateResponse{}

	armTemplates := arm.NewTemplates(arm.Params{
		Cfg: a.cfg,
	})
	tmplByts, checksum, err := armTemplates.Template(&req.Input)
	if err != nil {
		return res, temporal.NewNonRetryableApplicationError(
			"unable to create ARM template",
			"arm_template_error",
			err,
		)
	}

	return RenderARMStackTemplateResponse{
		RAWJson:  tmplByts,
		Checksum: checksum,
	}, nil
}
