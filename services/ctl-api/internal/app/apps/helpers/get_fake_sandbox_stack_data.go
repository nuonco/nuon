package helpers

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetFakeSandboxStackData builds a fake stack output map for sandbox mode orgs,
// using real keys from the app config with random fake values.
// stateMap is used to render templated role names (e.g. "{{.install.name}}-admin").
// NOTE: this is a standalone function because it is used directly within workflows.
func GetFakeSandboxStackData(appCfg *app.AppConfig, region string, stateMap map[string]interface{}) map[string]any {
	breakGlassRoleARNs := make(map[string]string)
	for _, role := range appCfg.BreakGlassConfig.Roles {
		name := renderName(role.Name, stateMap)
		breakGlassRoleARNs[name] = fmt.Sprintf("arn:aws:iam::123456789012:role/%s-fake-%s", name, generics.GetFakeObj[string]())
	}

	customRoleARNs := make(map[string]string)
	for _, role := range appCfg.PermissionsConfig.CustomRoles {
		name := renderName(role.Name, stateMap)
		customRoleARNs[name] = fmt.Sprintf("arn:aws:iam::123456789012:role/%s-fake-%s", name, generics.GetFakeObj[string]())
	}

	installInputs := make(map[string]string)
	for _, input := range appCfg.InputConfig.AppInputs {
		if input.Source == app.AppInputSourceCustomer {
			installInputs[input.Name] = fmt.Sprintf("fake-%s-%s", input.Name, generics.GetFakeObj[string]())
		}
	}

	stackRefs := GetStackReferences(appCfg)

	data := map[string]any{
		"account":                  generics.GetFakeObj[string](),
		"region":                   region,
		"url":                      generics.GetFakeObj[string](),
		"maintenance_iam_role_arn": generics.GetFakeObj[string](),
		"provision_iam_role_arn":   generics.GetFakeObj[string](),
		"deprovision_iam_role_arn": generics.GetFakeObj[string](),
		"reprovision_iam_role_arn": generics.GetFakeObj[string](),
		"break_glass_role_arns":    breakGlassRoleARNs,
		"install_inputs":           installInputs,
		"custom_role_arns":         customRoleARNs,
		"vpc_id":                   generics.GetFakeObj[string](),
		"account_id":               generics.GetFakeObj[string](),
		"public_subnets":           generics.GetFakeObj[string](),
		"private_subnets":          generics.GetFakeObj[string](),
		"runner_subnet":            generics.GetFakeObj[string](),
	}

	return generics.MergeMap(refs.GetFakeRefs(stackRefs), data)
}

// renderName renders a templated name using state, falling back to the raw name on error.
func renderName(name string, stateMap map[string]interface{}) string {
	if stateMap == nil {
		return name
	}
	rendered, err := render.RenderV2(name, stateMap)
	if err != nil {
		return name
	}
	return rendered
}
