package activities

import (
	"context"
	"fmt"
	"maps"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

type RenderAWSStackTemplateRequest struct {
	Input stacks.TemplateInput `temporaljson:"input"`
}

type RenderAWSStackTemplateResponse struct {
	RAWJson                        []byte                       `temporaljson:"raw_json"`
	Checksum                       string                       `temporaljson:"checksum"`
	AWSStackParameters             map[string]string            `temporaljson:"aws_stack_parameters,omitempty"`
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

	var parameters map[string]string
	if !req.Input.CustomStacksOnly && !req.Input.BundleCompile && req.Input.Install.InstallConfig != nil && len(req.Input.Install.InstallConfig.AWSStackParameters) > 0 {
		if req.Input.Install.CloudConnectionID == nil {
			return res, fmt.Errorf("root stack parameter overrides require an AWS cloud connection")
		}
		parameters = maps.Clone(req.Input.Install.InstallConfig.AWSStackParameters)
		for name := range parameters {
			if _, ok := tmpl.Parameters[name]; !ok {
				return res, fmt.Errorf("root CloudFormation parameter %q is not declared in the install stack template", name)
			}
		}
		stateData, err := req.Input.InstallState.AsMap()
		if err != nil {
			return res, errors.Wrap(err, "unable to get install state for stack parameters")
		}
		if err := render.RenderTextStringMap(parameters, stateData); err != nil {
			return res, errors.Wrap(err, "unable to render root stack parameters")
		}
	}

	// Must run before tmpl.JSON(): strips the side-channel metadata key so it
	// never reaches the deployed template.
	inputParametersMap := cloudformation.ExtractAndStripCustomStacksInputParameters(tmpl)

	tmplByts, err := tmpl.JSON()
	if err != nil {
		return res, errors.Wrap(err, "unable to get cloudformation json")
	}

	res = RenderAWSStackTemplateResponse{
		Checksum:           awsChecksum,
		RAWJson:            tmplByts,
		AWSStackParameters: parameters,
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
