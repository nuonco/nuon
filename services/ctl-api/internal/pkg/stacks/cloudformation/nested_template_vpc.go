package cloudformation

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"
	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// these parameter values should never be provided by the customer. these are always provided by nuon.
// as a result, if they are specified as input Params, we ensure the values are configurable only by us
// by excluding them from the Parameters for the nested stack.
var ReservedParamNames = []string{"ClusterName", "Namespaces", "NuonInstallID", "NuonAppID", "NuonOrgID"}

func (tpl *Templates) getClusterName(inp *stacks.TemplateInput) string {
	// NOTE: we need to provide a cluster name to the eks vpc template in order to pre-tag the subnets
	// historically, we've pre-emptively used the install.id. this is not necessarily the best
	// because it's forced us to re-tag subnets in the sandbox. that's fine but we prefer it to be
	// more correct. this method attempts to fetch a cluster_name from the inputs and uses that, if
	// present.
	// NOTE: this decision is essentially encoding a convention but this is fine because we consider mapping
	// one well-known input to a parameter to be uncontroversial.

	if inp.Install.CurrentInstallInputs != nil {
		if val, ok := inp.Install.CurrentInstallInputs.Values["cluster_name"]; ok && val != nil {
			return *val
		}
	}

	return inp.Install.ID
}

// VPCNestedStack returns a nested stack template for VPC resources
func (tpl *Templates) getVPCNestedStack(inp *stacks.TemplateInput, t tagBuilder) (*nestedcloudformation.Stack, map[string]cloudformation.Parameter, error) {
	parameters, defaultParameters, reservedInTemplate, _, err := tpl.extractNestedStackParameters(inp.AppCfg.StackConfig.VPCNestedTemplateURL)
	if err != nil {
		return nil, nil, fmt.Errorf("VPC nested stack: %w", err)
	}

	// these common params should always be set by nuon so they are explicitly
	// set aside so they can override any template values
	commonParams := map[string]string{
		"NuonInstallID": inp.Install.ID,
		"NuonAppID":     inp.Install.AppID,
		"NuonOrgID":     inp.Install.OrgID,
	}

	// only include ClusterName if the nested template defines it as a parameter,
	// otherwise CloudFormation will fail with an unknown parameter error
	if reservedInTemplate["ClusterName"] {
		commonParams["ClusterName"] = tpl.getClusterName(inp)
	}

	// merge specific params into common params
	maps.Copy(parameters, commonParams)

	stack := &nestedcloudformation.Stack{
		Parameters: parameters,
		TemplateURL: cloudformation.Join("", []any{
			inp.AppCfg.StackConfig.VPCNestedTemplateURL,
		}),
		Tags: t.apply(nil, "vpc"),
	}

	return stack, defaultParameters, nil
}

type cfnTemplateShape struct {
	Parameters map[string]struct {
		Type        string      `yaml:"Type" json:"Type"`
		Description string      `yaml:"Description" json:"Description,omitempty"`
		Default     interface{} `yaml:"Default" json:"Default,omitempty"`
	} `yaml:"Parameters"`
	Outputs map[string]struct{} `yaml:"Outputs"`
}

func (tpl *Templates) fetchTemplate(templateURL string) (*cfnTemplateShape, error) {
	resp, err := http.Get(templateURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, templateURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tmpl cfnTemplateShape
	if strings.HasSuffix(templateURL, ".json") {
		err = json.Unmarshal(body, &tmpl)
	} else {
		err = yaml.Unmarshal(body, &tmpl)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &tmpl, nil
}

func (tpl *Templates) extractNestedStackParameters(templateURL string, additionalReservedNames ...string) (map[string]string, map[string]cloudformation.Parameter, map[string]bool, map[string]struct{}, error) {
	params := map[string]string{}
	defaultParams := map[string]cloudformation.Parameter{}
	reservedInTemplate := map[string]bool{}
	templateOutputs := map[string]struct{}{}

	tmpl, err := tpl.fetchTemplate(templateURL)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch template %s: %w", templateURL, err)
	}

	for paramName, param := range tmpl.Parameters {
		if slices.Contains(ReservedParamNames, paramName) || slices.Contains(additionalReservedNames, paramName) {
			reservedInTemplate[paramName] = true
			continue
		}

		params[paramName] = cloudformation.Ref(paramName)

		cfnParam := cloudformation.Parameter{
			Type:    param.Type,
			Default: param.Default,
		}
		if param.Description != "" {
			desc := param.Description
			cfnParam.Description = &desc
		}
		defaultParams[paramName] = cfnParam
	}

	templateOutputs = tmpl.Outputs
	return params, defaultParams, reservedInTemplate, templateOutputs, nil
}

type firstClassOutput struct {
	resource   string
	outputName string
}

func (tpl *Templates) extractFirstClassOutputs(templateURLs map[string]string) map[string]firstClassOutput {
	outputs := map[string]firstClassOutput{}
	for resourceName, templateURL := range templateURLs {
		if templateURL == "" {
			continue
		}
		tmpl, err := tpl.fetchTemplate(templateURL)
		if err != nil {
			continue
		}
		for outputName := range tmpl.Outputs {
			outputs[outputName] = firstClassOutput{
				resource:   resourceName,
				outputName: outputName,
			}
		}
	}
	return outputs
}
