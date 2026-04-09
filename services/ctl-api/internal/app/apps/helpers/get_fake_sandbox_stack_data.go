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
	installInputs := make(map[string]string)
	for _, input := range appCfg.InputConfig.AppInputs {
		if input.Source == app.AppInputSourceCustomer {
			installInputs[input.Name] = fmt.Sprintf("fake-%s-%s", input.Name, generics.GetFakeObj[string]())
		}
	}

	stackRefs := GetStackReferences(appCfg)

	var data map[string]any

	switch appCfg.RunnerConfig.Type {
	case app.AppRunnerTypeGCP:
		breakGlassSAEmails := make(map[string]string)
		for _, role := range appCfg.BreakGlassConfig.Roles {
			name := renderName(role.Name, stateMap)
			breakGlassSAEmails[name] = fmt.Sprintf("%s-fake-%s@fake-project.iam.gserviceaccount.com", name, generics.GetFakeObj[string]())
		}

		customSAEmails := make(map[string]string)
		for _, role := range appCfg.PermissionsConfig.CustomRoles {
			name := renderName(role.Name, stateMap)
			customSAEmails[name] = fmt.Sprintf("%s-fake-%s@fake-project.iam.gserviceaccount.com", name, generics.GetFakeObj[string]())
		}

		fakeProject := fmt.Sprintf("fake-project-%s", generics.GetFakeObj[string]())
		data = map[string]any{
			"project_id":                   fakeProject,
			"region":                       region,
			"network_name":                 fmt.Sprintf("vpc-%s", generics.GetFakeObj[string]()),
			"network_id":                   fmt.Sprintf("projects/%s/global/networks/vpc-%s", fakeProject, generics.GetFakeObj[string]()),
			"public_subnet_name":           fmt.Sprintf("public-%s", generics.GetFakeObj[string]()),
			"private_subnet_name":          fmt.Sprintf("private-%s", generics.GetFakeObj[string]()),
			"runner_subnet_name":           fmt.Sprintf("runner-%s", generics.GetFakeObj[string]()),
			"runner_service_account_email": fmt.Sprintf("runner-%s@%s.iam.gserviceaccount.com", generics.GetFakeObj[string](), fakeProject),
			"gke_node_pool_sa_email":       fmt.Sprintf("gke-nodes-%s@%s.iam.gserviceaccount.com", generics.GetFakeObj[string](), fakeProject),
			"provision_sa_email":           fmt.Sprintf("provision-%s@%s.iam.gserviceaccount.com", generics.GetFakeObj[string](), fakeProject),
			"deprovision_sa_email":         fmt.Sprintf("deprovision-%s@%s.iam.gserviceaccount.com", generics.GetFakeObj[string](), fakeProject),
			"maintenance_sa_email":         fmt.Sprintf("maintenance-%s@%s.iam.gserviceaccount.com", generics.GetFakeObj[string](), fakeProject),
			"break_glass_sa_emails":        breakGlassSAEmails,
			"custom_sa_emails":             customSAEmails,
			"install_inputs":               installInputs,
			"secret_names":                 map[string]string{},
		}

	default: // AWS and everything else
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

		fakeAccountID := fmt.Sprintf("%012d", 100000000000+len(generics.GetFakeObj[string]()))
		data = map[string]any{
			"account":                  fakeAccountID,
			"account_id":               fakeAccountID,
			"region":                   region,
			"url":                      fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com", generics.GetFakeObj[string](), region),
			"maintenance_iam_role_arn": fmt.Sprintf("arn:aws:iam::%s:role/maintenance-%s", fakeAccountID, generics.GetFakeObj[string]()),
			"provision_iam_role_arn":   fmt.Sprintf("arn:aws:iam::%s:role/provision-%s", fakeAccountID, generics.GetFakeObj[string]()),
			"deprovision_iam_role_arn": fmt.Sprintf("arn:aws:iam::%s:role/deprovision-%s", fakeAccountID, generics.GetFakeObj[string]()),
			"reprovision_iam_role_arn": fmt.Sprintf("arn:aws:iam::%s:role/reprovision-%s", fakeAccountID, generics.GetFakeObj[string]()),
			"runner_iam_role_arn":      fmt.Sprintf("arn:aws:iam::%s:role/runner-%s", fakeAccountID, generics.GetFakeObj[string]()),
			"break_glass_role_arns":    breakGlassRoleARNs,
			"install_inputs":           installInputs,
			"custom_role_arns":         customRoleARNs,
			"vpc_id":                   fmt.Sprintf("vpc-%s", generics.GetFakeObj[string]()),
			"public_subnets":           fmt.Sprintf("subnet-%s,subnet-%s", generics.GetFakeObj[string](), generics.GetFakeObj[string]()),
			"private_subnets":          fmt.Sprintf("subnet-%s,subnet-%s", generics.GetFakeObj[string](), generics.GetFakeObj[string]()),
			"runner_subnet":            fmt.Sprintf("subnet-%s", generics.GetFakeObj[string]()),
			"github_app_key_arn":       generics.GetFakeObj[string](),
		}
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
