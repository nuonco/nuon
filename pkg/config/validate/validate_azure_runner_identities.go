package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/config"
)

type armParametersShape struct {
	Parameters map[string]json.RawMessage `yaml:"parameters" json:"parameters"`
}

// ValidateAzureRunnerIdentities guards the one combination that otherwise fails
// with an opaque IMDS "Identity not found" at runtime: an azure-bicep app that
// defines per-operation identities (any azure permission role) AND pins a custom
// runner_nested_template_url whose template does not accept a
// "userAssignedIdentities" parameter. Only the built-in runner (empty URL) or a
// template that declares that parameter can attach the identities to the VMSS.
func ValidateAzureRunnerIdentities(a *config.AppConfig) error {
	if a.Stack == nil || a.Stack.Type != "azure-bicep" || a.Stack.RunnerNestedTemplateURL == "" {
		return nil
	}
	if a.Permissions == nil || !hasAzurePermissionRole(a.Permissions) {
		return nil
	}

	params, err := fetchTemplateParameters(a.Stack.RunnerNestedTemplateURL)
	if err != nil {
		// Don't fail validation on a transient fetch error; the render-time
		// check in ctl-api still enforces this.
		return nil
	}
	if _, ok := params["userAssignedIdentities"]; !ok {
		return fmt.Errorf(
			"custom runner_nested_template_url %q must declare a 'userAssignedIdentities' parameter to attach this app's per-operation Azure managed identities; omit runner_nested_template_url to use the built-in runner",
			a.Stack.RunnerNestedTemplateURL,
		)
	}
	return nil
}

func hasAzurePermissionRole(p *config.PermissionsConfig) bool {
	roles := []*config.AppAWSIAMRole{p.ProvisionRole, p.DeprovisionRole, p.MaintenanceRole}
	roles = append(roles, p.CustomRoles...)
	roles = append(roles, p.Roles...)
	for _, r := range roles {
		if r != nil && r.CloudPlatform == "azure" {
			return true
		}
	}
	return false
}

func fetchTemplateParameters(templateURL string) (map[string]json.RawMessage, error) {
	resp, err := templateOutputsClient.Get(templateURL)
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
	var tmpl armParametersShape
	if strings.HasSuffix(templateURL, ".json") {
		err = json.Unmarshal(body, &tmpl)
	} else {
		err = yaml.Unmarshal(body, &tmpl)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse template parameters: %w", err)
	}
	return tmpl.Parameters, nil
}
