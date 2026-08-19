package arm

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func secretsTemplateInput(scope app.StackDeploymentScope) *stacks.TemplateInput {
	inp := azureRolesTemplateInput()
	inp.DeploymentScope = scope
	inp.AppCfg.SecretsConfig.Secrets = []app.AppSecretConfig{
		{Name: "db_password", Description: "Database password.", Required: true},
		{Name: "api_token", Default: "prefilled", Description: "API token."},
		{Name: "generated_key", AutoGenerate: true},
	}
	return inp
}

func TestKeyVault_SecretsBecomeSecureStringParameters(t *testing.T) {
	inp := secretsTemplateInput(app.StackDeploymentScopeSubscription)
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatal(err)
	}

	pw, ok := armTmpl.Parameters["secretDbPassword"]
	if !ok {
		t.Fatalf("db_password not surfaced; parameters: %v", sortedKeys(armTmpl.Parameters))
	}
	if pw.Type != "securestring" {
		t.Errorf("secretDbPassword type = %q, want securestring — a plain string puts the value in the deployment history", pw.Type)
	}
	// No default, so ARM refuses to deploy without one rather than writing a blank
	// secret over a real one.
	if pw.DefaultValue != nil {
		t.Errorf("a required secret must not get a default, got %#v", pw.DefaultValue)
	}
	if pw.Metadata == nil || pw.Metadata.Description != "Database password." {
		t.Errorf("description not carried to the portal form: %+v", pw.Metadata)
	}

	tok := armTmpl.Parameters["secretApiToken"]
	if tok.DefaultValue != "prefilled" {
		t.Errorf("secretApiToken default = %#v, want the configured default", tok.DefaultValue)
	}

	// Nothing on the Azure path generates auto-generate secrets, so asking the
	// customer for one would be asking for a value they do not have.
	if _, ok := armTmpl.Parameters["secretGeneratedKey"]; ok {
		t.Error("auto-generate secret surfaced as a customer input")
	}
}

func TestKeyVault_VaultAndSecretsCreatedInTheInstallGroup(t *testing.T) {
	inp := secretsTemplateInput(app.StackDeploymentScopeSubscription)
	tmpl := &Templates{cfg: &internal.Config{}}

	res := tmpl.getKeyVaultResources(inp, scopeFor(inp))
	if len(res) != 1 {
		t.Fatalf("expected a single wrapper, got %d", len(res))
	}
	wrapper := res[0].(map[string]any)

	if got := wrapper["resourceGroup"]; got != "[variables('installResourceGroupName')]" {
		t.Errorf("wrapper resourceGroup = %v", got)
	}
	props := wrapper["properties"].(map[string]any)
	if got := props["expressionEvaluationOptions"].(map[string]any)["scope"]; got != "inner" {
		t.Errorf("a securestring cannot cross an outer-evaluation boundary, got scope %v", got)
	}

	// The value must be securestring on both sides of the wrapper.
	innerDecl := props["template"].(map[string]any)["parameters"].(map[string]any)
	for _, name := range []string{"secretDbPassword", "secretApiToken"} {
		p, ok := innerDecl[name].(map[string]any)
		if !ok {
			t.Errorf("inner template does not declare %q", name)
			continue
		}
		if p["type"] != "securestring" {
			t.Errorf("inner %s type = %v, want securestring", name, p["type"])
		}
	}

	inner := innerResources(t, wrapper)
	if got := countResourceType(inner, "Microsoft.KeyVault/vaults"); got != 1 {
		t.Errorf("expected exactly one vault, got %d", got)
	}
	if got := countResourceType(inner, "Microsoft.KeyVault/vaults/secrets"); got != 2 {
		t.Errorf("expected 2 customer secrets, got %d", got)
	}

	var vault map[string]any
	secretNames := map[string]string{}
	for _, r := range inner {
		m := r.(map[string]any)
		switch m["type"] {
		case "Microsoft.KeyVault/vaults":
			vault = m
		case "Microsoft.KeyVault/vaults/secrets":
			secretNames[m["name"].(string)] = m["properties"].(map[string]any)["value"].(string)
		}
	}

	// RBAC, matching how the runner's role assignment grants access; access policies
	// would leave the runner unable to read despite the assignment.
	if got := vault["properties"].(map[string]any)["enableRbacAuthorization"]; got != true {
		t.Errorf("enableRbacAuthorization = %v", got)
	}

	// Underscores are illegal in Key Vault secret names and the phone-home builds each
	// URI from the same mapping, so these have to agree.
	for name, value := range secretNames {
		if strings.Contains(name, "_") {
			t.Errorf("secret name %q keeps an underscore", name)
		}
		if !strings.HasPrefix(value, "[parameters('secret") {
			t.Errorf("secret %q value is not read from its parameter: %s", name, value)
		}
	}
	if _, ok := secretNames["[format('{0}/db-password', take(format('{0}', parameters('nuonInstallID')), 24))]"]; !ok {
		t.Errorf("db_password secret not named by convention, got %v", sortedKeys(secretNames))
	}
}

// An app with no secrets still needs the vault: the runner is granted a role on it
// and the phone-home reports its ID, so its absence fails the deploy regardless.
func TestKeyVault_CreatedEvenWithNoSecrets(t *testing.T) {
	inp := subscriptionTemplateInput()
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatal(err)
	}

	var wrapper map[string]any
	for _, r := range armTmpl.Resources {
		if m := r.(map[string]any); m["name"] == keyVaultDeploymentName {
			wrapper = m
		}
	}
	if wrapper == nil {
		t.Fatal("no keyVaultDeployment rendered")
	}
	inner := innerResources(t, wrapper)
	if got := countResourceType(inner, "Microsoft.KeyVault/vaults"); got != 1 {
		t.Errorf("expected the vault, got %d", got)
	}
	if got := countResourceType(inner, "Microsoft.KeyVault/vaults/secrets"); got != 0 {
		t.Errorf("expected no secrets, got %d", got)
	}

	// No secrets means no new portal fields.
	for name := range armTmpl.Parameters {
		if strings.HasPrefix(name, "secret") {
			t.Errorf("unexpected secret parameter %q", name)
		}
	}
}

// At resource-group scope the group already exists when the customer deploys, so the
// vault stays a documented prerequisite and the template must not change at all.
func TestKeyVault_NotCreatedAtResourceGroupScope(t *testing.T) {
	inp := secretsTemplateInput(app.StackDeploymentScopeResourceGroup)
	tmpl := &Templates{cfg: &internal.Config{}}

	if got := tmpl.getKeyVaultResources(inp, scopeFor(inp)); got != nil {
		t.Errorf("resource-group scope must not create the vault, got %d resources", len(got))
	}
	if got := azureSecretParameters(inp, scopeFor(inp)); got != nil {
		t.Errorf("resource-group scope must not add secret parameters, got %v", sortedKeys(got))
	}
}

// The runner is granted a role on the vault, and the phone-home reports its ID; both
// fail if they run first.
func TestKeyVault_OrderedBeforeItsConsumers(t *testing.T) {
	inp := secretsTemplateInput(app.StackDeploymentScopeSubscription)
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatal(err)
	}

	deps := map[string][]string{}
	for _, r := range armTmpl.Resources {
		m := r.(map[string]any)
		name, _ := m["name"].(string)
		switch d := m["dependsOn"].(type) {
		case []string:
			deps[name] = d
		case []any:
			for _, v := range d {
				if s, ok := v.(string); ok {
					deps[name] = append(deps[name], s)
				}
			}
		}
	}

	for _, consumer := range []string{runnerGrantsDeploymentName, phoneHomeDeploymentName} {
		if !slices.Contains(deps[consumer], keyVaultDeploymentName) {
			t.Errorf("%s does not depend on %s, got %v", consumer, keyVaultDeploymentName, deps[consumer])
		}
	}
}

func TestAzureSecretParamName(t *testing.T) {
	for in, want := range map[string]string{
		"db_password":     "secretDbPassword",
		"api_token":       "secretApiToken",
		"license":         "secretLicense",
		"smtp_host_2":     "secretSmtpHost2",
		"oidc-client-id":  "secretOidcClientId",
		"already.dotted":  "secretAlreadyDotted",
		"keepsInnerCaseX": "secretKeepsInnerCaseX",
	} {
		if got := azureSecretParamName(in); got != want {
			t.Errorf("azureSecretParamName(%q) = %q, want %q", in, got, want)
		}
	}
}

// Writes the rendered subscription-scope template so it can be validated against a
// real subscription: AZURE_RENDER_OUT=/path go test -run TestKeyVault_RenderForAzureValidation
func TestKeyVault_RenderForAzureValidation(t *testing.T) {
	out := os.Getenv("AZURE_RENDER_OUT")
	if out == "" {
		t.Skip("AZURE_RENDER_OUT not set")
	}
	tmpl := &Templates{cfg: &internal.Config{}}
	armTmpl, err := tmpl.getAzureTemplate(secretsTemplateInput(app.StackDeploymentScopeSubscription))
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(armTmpl, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, b, 0o644); err != nil {
		t.Fatal(err)
	}
}
