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

	// QuickLinkWrapperJSON is the template behind the Azure portal quick link, and
	// is empty unless the stack version carries a QuickLinkBucketKey. It is a
	// separate artifact because the portal can only create a plain deployment —
	// see arm.QuickLinkWrapper.
	QuickLinkWrapperJSON     []byte `temporaljson:"quick_link_wrapper_json"`
	QuickLinkWrapperChecksum string `temporaljson:"quick_link_wrapper_checksum"`

	// QuickLinkUIDefJSON is the createUiDefinition the quick link pairs with the
	// wrapper. It constrains the portal's Basics step so a reprovision updates the
	// install's stack rather than creating a second one in another resource group.
	QuickLinkUIDefJSON     []byte `temporaljson:"quick_link_ui_def_json"`
	QuickLinkUIDefChecksum string `temporaljson:"quick_link_ui_def_checksum"`
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

	res.RAWJson = tmplByts
	res.Checksum = checksum

	stackVersion := req.Input.CloudFormationStackVersion
	if stackVersion == nil || stackVersion.QuickLinkBucketKey == "" {
		return res, nil
	}

	wrapperByts, wrapperChecksum, err := armTemplates.QuickLinkWrapper(&req.Input, stackVersion.TemplateURL)
	if err != nil {
		return res, temporal.NewNonRetryableApplicationError(
			"unable to create ARM quick link wrapper",
			"arm_template_error",
			err,
		)
	}
	res.QuickLinkWrapperJSON = wrapperByts
	res.QuickLinkWrapperChecksum = wrapperChecksum

	if stackVersion.QuickLinkUIDefBucketKey == "" {
		return res, nil
	}

	uiDefByts, uiDefChecksum, err := armTemplates.QuickLinkUIDefinition(&req.Input)
	if err != nil {
		return res, temporal.NewNonRetryableApplicationError(
			"unable to create ARM quick link UI definition",
			"arm_template_error",
			err,
		)
	}
	res.QuickLinkUIDefJSON = uiDefByts
	res.QuickLinkUIDefChecksum = uiDefChecksum

	return res, nil
}
