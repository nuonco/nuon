package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/pkg/errors"
)

type RenderAWSStackTemplateRequest struct {
	Input stacks.TemplateInput `temporaljson:"input"`
}

type RenderAWSStackTemplateResponse struct {
	RAWJson                        []byte                       `temporaljson:"raw_json"`
	Checksum                       string                       `temporaljson:"checksum"`
	CustomStacksOutputMap          map[string]map[string]string `temporaljson:"custom_stacks_output_map"`
	CustomStacksInputParametersMap map[string]map[string]string `temporaljson:"custom_stacks_input_parameters_map"`
}

// @temporal-gen-v2 activity
func (a *Activities) RenderAWSStackTemplate(ctx context.Context, req *RenderAWSStackTemplateRequest) (RenderAWSStackTemplateResponse, error) {
	res := RenderAWSStackTemplateResponse{}
	tmpl, awsChecksum, err := a.cfTemplates.Template(&req.Input)
	if err != nil {
		return res, errors.Wrap(err, "unable to create cloudformation template")
	}

	// Must run before tmpl.JSON(): strips the side-channel metadata key so it
	// never reaches the deployed template.
	inputParametersMap := cloudformation.ExtractAndStripCustomStacksInputParameters(tmpl)

	tmplByts, err := tmpl.JSON()
	if err != nil {
		return res, errors.Wrap(err, "unable to get cloudformation json")
	}

	res = RenderAWSStackTemplateResponse{
		Checksum: awsChecksum,
		RAWJson:  tmplByts,
	}

	if req.Input.CustomStacksOnly {
		flatOutputs := make(map[string]string, len(tmpl.Outputs))
		for name := range tmpl.Outputs {
			flatOutputs[name] = name
		}
		allStackNames := make([]string, 0, len(req.Input.AppCfg.StackConfig.CustomNestedStacks))
		for _, stack := range req.Input.AppCfg.StackConfig.CustomNestedStacks {
			allStackNames = append(allStackNames, stack.Name)
		}

		outputMap := map[string]map[string]string{}
		for stackName, stackResult := range cloudformation.SplitCustomStacksOnlyOutputs(flatOutputs, allStackNames) {
			outputMap[stackName] = stackResult["outputs"]
		}
		res.CustomStacksOutputMap = outputMap
		res.CustomStacksInputParametersMap = inputParametersMap
	}

	return res, nil
}
