package arm

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// keyVaultDeploymentName is the nested deployment the install Key Vault and its
// secrets are created in at subscription scope.
const keyVaultDeploymentName = "keyVaultDeployment"

const keyVaultAPIVersion = "2023-07-01"

// azureKeyVaultSecretName is the sole source of truth for how an app-config secret
// name maps onto a Key Vault secret name — Key Vault allows only alphanumerics and
// hyphens. The phone-home builds each secret's URI from the same mapping, so the two
// must not drift.
func azureKeyVaultSecretName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

// azureSecretParamName turns an app-config secret name into an ARM parameter name.
// The portal derives its form label from the parameter name, so this is camelCased
// rather than carrying the config's underscores through: db_password reads as
// "Secret Db Password".
func azureSecretParamName(name string) string {
	out := "secret"
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		r := []rune(part)
		out += strings.ToUpper(string(r[0])) + string(r[1:])
	}
	return out
}

type azureSecret struct {
	name      string
	kvName    string
	paramName string
	// description is what the customer reads next to the field in the portal.
	description string
	// defaultValue empty means the customer has to supply one; ARM then refuses to
	// deploy without it rather than writing a blank secret.
	defaultValue string
}

// azureCustomerSecrets are the secrets whose values the customer supplies. Sorted so
// the render is deterministic.
//
// AutoGenerate secrets are excluded because nothing on the Azure path generates them
// — they are unimplemented rather than handled elsewhere.
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

// azureSecretParameters surfaces each customer-supplied secret as a securestring in
// the root, which the portal renders as a masked field on the deployment form.
//
// Only at subscription scope. At resource-group scope the customer creates the vault
// and its secrets by hand before deploying, because the resource group already
// exists by then.
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

// getKeyVaultResources creates the install Key Vault and the customer's secrets in
// the install resource group.
//
// This exists because subscription scope removed the customer's `az group create`
// step, which they used to run before `az keyvault create`. With the resource group
// now created by the stack itself, there is no longer a point at which the customer
// could have made the vault by hand — the deploy would fail assigning the runner a
// role on a vault that cannot exist yet.
//
// Returns nil at resource-group scope, where the vault stays a documented
// prerequisite and the rendered template is unchanged.
func (t *Templates) getKeyVaultResources(inp *stacks.TemplateInput, scope armScope) []any {
	if !scope.subscription {
		return nil
	}

	// Read at resource-group scope: these expressions live inside the wrapper, where
	// the vault's own resource group is the ambient one.
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
				"sku":      map[string]any{"family": "A", "name": "standard"},
				"tenantId": "[subscription().tenantId]",
				// RBAC rather than access policies, matching how the runner's role
				// assignment grants access.
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
		// securestring the whole way down: a secure value cannot cross into a nested
		// deployment that uses outer evaluation, and declaring it as a plain string
		// here would put the value in the deployment history.
		params[s.paramName] = nestedParam{
			typ:   "securestring",
			value: fmt.Sprintf("[parameters('%s')]", s.paramName),
		}
	}

	return scope.wrapInInstallRG(keyVaultDeploymentName, params, resources, nil)
}
