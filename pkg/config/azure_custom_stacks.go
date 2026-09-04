package config

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

const bicepFileExtension = ".bicep"

// ValidateAzureCustomNestedStacks rejects azure-bicep custom nested stacks whose
// template is Bicep source instead of the compiled ARM JSON that ARM's
// templateLink can actually fetch. Called from the CLI validation path and from
// the server-side config builder, so a config that bypasses the CLI — a direct
// API sync or a VCS branch run — is rejected the same way.
//
// Contents is the template body resolved from template_url; it is empty on call
// paths that have not fetched it yet, and the check is skipped there.
func ValidateAzureCustomNestedStacks(stackType string, stacks []CustomNestedStack) error {
	if stackType != "azure-bicep" {
		return nil
	}

	for i, stack := range stacks {
		if strings.HasSuffix(strings.ToLower(stack.TemplateURL), bicepFileExtension) {
			return newAzureBicepSourceErr(i, stack, fmt.Sprintf(
				"template_url %q is Bicep source",
				stack.TemplateURL,
			))
		}

		if strings.TrimSpace(stack.Contents) == "" {
			continue
		}

		if !isARMJSON(stack.Contents) {
			return newAzureBicepSourceErr(i, stack, fmt.Sprintf(
				"the contents of template_url %q are not valid ARM JSON",
				stack.TemplateURL,
			))
		}
	}

	return nil
}

func newAzureBicepSourceErr(index int, stack CustomNestedStack, problem string) error {
	msg := fmt.Sprintf(
		"custom_nested_stacks[%d] (%s): azure-bicep custom stacks must reference compiled ARM JSON, but %s. Compile it and point template_url at the output:\n\n    %s",
		index, stack.Name, problem, bicepBuildCommand(stack.TemplateURL),
	)

	return ErrConfig{
		Description: msg,
		Err:         fmt.Errorf("%s", msg),
	}
}

func isARMJSON(contents string) bool {
	var tmpl struct {
		Schema         string            `json:"$schema"`
		ContentVersion string            `json:"contentVersion"`
		Resources      []json.RawMessage `json:"resources"`
	}
	return json.Unmarshal([]byte(contents), &tmpl) == nil &&
		tmpl.Schema != "" &&
		tmpl.ContentVersion != "" &&
		tmpl.Resources != nil
}

func bicepBuildCommand(templateURL string) string {
	base := strings.TrimSuffix(templateURL, path.Ext(templateURL))
	source := templateURL
	outfile := base + ".json"
	if !strings.EqualFold(path.Ext(templateURL), bicepFileExtension) {
		source = base + bicepFileExtension
		outfile = templateURL
	}

	return fmt.Sprintf("az bicep build --file %s --outfile %s", source, outfile)
}
