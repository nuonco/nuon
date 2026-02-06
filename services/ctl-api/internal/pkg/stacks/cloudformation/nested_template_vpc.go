package cloudformation

import (
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/awslabs/goformation/v7"
	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// these parameter values should never be provided by the customer. these are always provided by nuon.
// as a result, if they are specified as input Params, we ensure the values are configurable only by us
// by excluding them from the Parameters for the nested stack.
var ReservedParamNames = []string{"ClusterName", "NuonInstallID", "NuonAppID", "NuonOrgID"}

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
func (tpl *Templates) getVPCNestedStack(inp *stacks.TemplateInput, t tagBuilder) (*nestedcloudformation.Stack, map[string]cloudformation.Parameter) {
	parameters, defaultParameters, reservedInTemplate := tpl.extractNestedStackParameters(inp.AppCfg.StackConfig.VPCNestedTemplateURL)

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

	return stack, defaultParameters

}

func (tpl *Templates) extractNestedStackParameters(templateURL string) (map[string]string, map[string]cloudformation.Parameter, map[string]bool) {
	params := map[string]string{}
	defaultParams := map[string]cloudformation.Parameter{}
	reservedInTemplate := map[string]bool{}

	resp, err := http.Get(templateURL)
	if err != nil {
		return params, defaultParams, reservedInTemplate
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return params, defaultParams, reservedInTemplate
	}

	// TODO(fd): pretty sure we only support yaml atm
	var tmpl *cloudformation.Template
	if strings.HasSuffix(templateURL, ".json") {
		tmpl, err = goformation.ParseJSON(body)
	} else {
		tmpl, err = goformation.ParseYAML(body)
	}
	if err != nil {
		return params, defaultParams, reservedInTemplate
	}

	for paramName, param := range tmpl.Parameters {
		if slices.Contains(ReservedParamNames, paramName) {
			reservedInTemplate[paramName] = true
			continue
		}

		params[paramName] = cloudformation.Ref(paramName)

		defaultParams[paramName] = cloudformation.Parameter{
			Type:        param.Type,
			Description: param.Description,
			Default:     param.Default,
		}
	}

	return params, defaultParams, reservedInTemplate
}
