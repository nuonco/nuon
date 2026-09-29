package arm

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const keyVaultDeploymentName = "keyVaultDeployment"

const keyVaultAPIVersion = "2023-07-01"

// why: azureKeyVaultSecretName is the sole source of truth for how an app-config secret
// name maps onto a Key Vault secret name — Key Vault allows only alphanumerics and
// hyphens. The phone-home builds each secret's URI from the same mapping, so the two
// must not drift.
func azureKeyVaultSecretName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

func azureSecretParamName(name string) string {
	return camelParamName("secret", name)
}

type azureSecret struct {
	name         string
	kvName       string
	paramName    string
	description  string
	defaultValue string
}

func azureCustomerSecrets(appCfg *app.AppConfig) []azureSecret {
	if appCfg == nil {
		return nil
	}

	var out []azureSecret
	for _, s := range appCfg.SecretsConfig.Secrets {
		if s.AutoGenerate {
			continue
		}
		desc := s.Description
		if desc == "" {
			desc = s.DisplayName
		}
		out = append(out, azureSecret{
			name:         s.Name,
			kvName:       azureKeyVaultSecretName(s.Name),
			paramName:    azureSecretParamName(s.Name),
			description:  desc,
			defaultValue: s.Default,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })

	return out
}

func azureSecretParameters(inp *stacks.TemplateInput, scope armScope) map[string]ARMParameter {
	if !scope.subscription {
		return nil
	}

	params := map[string]ARMParameter{}
	for _, s := range azureCustomerSecrets(inp.AppCfg) {
		p := ARMParameter{Type: "securestring"}
		if s.defaultValue != "" {
			p.DefaultValue = s.defaultValue
		}
		if s.description != "" {
			p.Metadata = &ARMParameterMetadata{Description: s.description}
		}
		params[s.paramName] = p
	}

	return params
}

func (t *Templates) getKeyVaultResources(inp *stacks.TemplateInput, scope armScope) []any {
	if !scope.subscription {
		return nil
	}

	inner := armScope{}
	vaultNameInner := inner.keyVaultNameInner()

	resources := []any{
		map[string]any{
			"type":       "Microsoft.KeyVault/vaults",
			"apiVersion": keyVaultAPIVersion,
			"name":       "[" + vaultNameInner + "]",
			"location":   "[parameters('location')]",
			"tags":       "[parameters('commonTags')]",
			"properties": map[string]any{
				"sku":                       map[string]any{"family": "A", "name": "standard"},
				"tenantId":                  "[subscription().tenantId]",
				"enableRbacAuthorization":   true,
				"enableSoftDelete":          true,
				"softDeleteRetentionInDays": 7,
			},
		},
	}

	params := map[string]nestedParam{
		"nuonInstallID": {typ: "string", value: scope.nuonIDRef("nuonInstallID")},
		"location":      {typ: "string", value: scope.rootLocationRef()},
		"commonTags":    {typ: "object", value: "[variables('commonTags')]"},
	}

	for _, s := range azureCustomerSecrets(inp.AppCfg) {
		resources = append(resources, map[string]any{
			"type":       "Microsoft.KeyVault/vaults/secrets",
			"apiVersion": keyVaultAPIVersion,
			"name":       fmt.Sprintf("[format('{0}/%s', %s)]", s.kvName, vaultNameInner),
			"dependsOn":  []string{inner.rgResourceIDExpr("Microsoft.KeyVault/vaults", vaultNameInner)},
			"properties": map[string]any{
				"value": fmt.Sprintf("[parameters('%s')]", s.paramName),
			},
		})
		// why: securestring the whole way down: a secure value cannot cross into a nested
		// deployment that uses outer evaluation, and declaring it as a plain string
		// here would put the value in the deployment history.
		params[s.paramName] = nestedParam{
			typ:   "securestring",
			value: fmt.Sprintf("[parameters('%s')]", s.paramName),
		}
	}

	return scope.wrapInInstallRG(keyVaultDeploymentName, params, resources, nil)
}
