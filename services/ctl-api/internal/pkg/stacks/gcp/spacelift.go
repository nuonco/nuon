package gcp

import (
	"bytes"
	"encoding/base64"
	"text/template"

	"github.com/pkg/errors"
)

// defaultSpaceliftTerraformVersion is the Terraform version pinned on generated
// Spacelift stacks/blueprints. Kept in lockstep with the install-stacks//gcp
// module's versions.tf constraint.
const defaultSpaceliftTerraformVersion = "1.9.0"

type spaceliftTemplateInput struct {
	InstallID        string
	TerraformVersion string
	InputsB64        string
	SecretsB64       string
}

func newSpaceliftTemplateInput(inputsTfvars, secretsTfvars, installID string) spaceliftTemplateInput {
	return spaceliftTemplateInput{
		InstallID:        installID,
		TerraformVersion: defaultSpaceliftTerraformVersion,
		InputsB64:        base64.StdEncoding.EncodeToString([]byte(inputsTfvars)),
		SecretsB64:       base64.StdEncoding.EncodeToString([]byte(secretsTfvars)),
	}
}

func renderSpaceliftAdminTF(inputsTfvars, secretsTfvars, installID string) (string, error) {
	return renderSpaceliftTemplate("gcp-spacelift-admin-tf", spaceliftAdminTfTmpl, inputsTfvars, secretsTfvars, installID)
}

func renderSpaceliftBlueprint(inputsTfvars, secretsTfvars, installID string) (string, error) {
	return renderSpaceliftTemplate("gcp-spacelift-blueprint", spaceliftBlueprintTmpl, inputsTfvars, secretsTfvars, installID)
}

func renderSpaceliftTemplate(name, tmpl, inputsTfvars, secretsTfvars, installID string) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse %s template", name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, newSpaceliftTemplateInput(inputsTfvars, secretsTfvars, installID)); err != nil {
		return "", errors.Wrapf(err, "unable to execute %s template", name)
	}

	return buf.String(), nil
}
