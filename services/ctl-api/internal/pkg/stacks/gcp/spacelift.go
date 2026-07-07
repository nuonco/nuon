package gcp

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// blueprintFuncs are template helpers for rendering Spacelift blueprint YAML.
// yamlStr emits a double-quoted YAML scalar so free-text values (input names,
// descriptions) that contain YAML-significant characters such as ": " don't get
// misparsed as nested mappings.
var blueprintFuncs = template.FuncMap{
	"yamlStr": strconv.Quote,
}

// defaultSpaceliftTerraformVersion is the Terraform version pinned on generated
// Spacelift stacks/blueprints. Satisfies the install-stacks//gcp module's
// versions.tf constraint (>= 1.5) and stays within the versions Spacelift's
// TERRAFORM_FOSS workflow tool exposes.
const defaultSpaceliftTerraformVersion = "1.5.7"

// blueprintInstallInputID / blueprintSecretID map a Nuon install-input or
// secret name to the blueprint input id it's exposed as. The prefixes keep the
// two namespaces from colliding when an input and a secret share a name.
func blueprintInstallInputID(name string) string { return "input_" + name }
func blueprintSecretID(name string) string       { return "secret_" + name }

type spaceliftAdminTemplateInput struct {
	InstallID        string
	TerraformVersion string
}

func renderSpaceliftAdminTF(installID string) (string, error) {
	t, err := template.New("gcp-spacelift-admin-tf").Parse(spaceliftAdminTfTmpl)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse gcp spacelift admin template")
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, spaceliftAdminTemplateInput{
		InstallID:        installID,
		TerraformVersion: defaultSpaceliftTerraformVersion,
	}); err != nil {
		return "", errors.Wrap(err, "unable to execute gcp spacelift admin template")
	}
	return buf.String(), nil
}

// spaceliftBlueprintData is the caller-facing input for the blueprint renderer.
type spaceliftBlueprintData struct {
	InstallID     string
	InputsTfvars  string
	SecretsTfvars string
	GCPProjectID  string
	GCPRegion     string
	InstallInputs []GCPInstallInputTemplateInput
	Secrets       []GCPSecretTemplateInput
}

type blueprintInput struct {
	ID          string
	Name        string
	Type        string
	Description string
	Default     string
}

type spaceliftBlueprintTemplateInput struct {
	InstallID        string
	TerraformVersion string
	Inputs           []blueprintInput
	InputsTfvars     string
	SecretsTfvars    string
}

func renderSpaceliftBlueprint(data spaceliftBlueprintData) (string, error) {
	inputs := []blueprintInput{
		{ID: "gcp_project_id", Name: "GCP project ID", Type: "short_text", Default: data.GCPProjectID},
		{ID: "gcp_region", Name: "GCP region", Type: "short_text", Default: data.GCPRegion},
	}
	for _, in := range data.InstallInputs {
		inputs = append(inputs, blueprintInput{
			ID:      blueprintInstallInputID(in.Name),
			Name:    in.Name,
			Type:    "short_text",
			Default: in.Default,
		})
	}
	for _, s := range data.Secrets {
		inputs = append(inputs, blueprintInput{
			ID:          blueprintSecretID(s.Name),
			Name:        s.Name,
			Type:        "secret",
			Description: s.Description,
			Default:     s.Default,
		})
	}

	t, err := template.New("gcp-spacelift-blueprint").Funcs(blueprintFuncs).Parse(spaceliftBlueprintTmpl)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse gcp spacelift blueprint template")
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, spaceliftBlueprintTemplateInput{
		InstallID:        data.InstallID,
		TerraformVersion: defaultSpaceliftTerraformVersion,
		Inputs:           inputs,
		InputsTfvars:     indentLines(data.InputsTfvars, 10),
		SecretsTfvars:    indentLines(data.SecretsTfvars, 10),
	}); err != nil {
		return "", errors.Wrap(err, "unable to execute gcp spacelift blueprint template")
	}
	return strings.TrimLeft(buf.String(), "\n"), nil
}

// indentLines prefixes every non-empty line of s with n spaces, for embedding
// tfvars as a YAML block scalar. The trailing newline is dropped so the block
// doesn't emit a trailing whitespace-only line.
func indentLines(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}
