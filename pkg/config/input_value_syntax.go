package config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	yaml "gopkg.in/yaml.v3"
)

// Dedicated structured input types. Unlike the scalar types (string, number,
// bool, list, json), these carry a whole document whose syntax we can validate
// early — before an install deploy — so malformed YAML/HCL is rejected at
// config-parse and input-update time rather than failing mid-deploy.
const (
	InputTypeYAML = "yaml"
	InputTypeHCL  = "hcl"
)

// ValidateInputValueSyntax checks that value is syntactically valid for the
// given input type. Empty values are always valid: an unset or cleared override
// is an exact deploy-time no-op. Scalar/simple types are not syntax-checked here
// and return nil.
func ValidateInputValueSyntax(inputType, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	switch inputType {
	case InputTypeYAML:
		var v interface{}
		if err := yaml.Unmarshal([]byte(value), &v); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		return nil
	case InputTypeHCL:
		parser := hclparse.NewParser()
		var (
			file  *hcl.File
			diags hcl.Diagnostics
		)
		if strings.HasPrefix(strings.TrimSpace(value), "{") {
			file, diags = parser.ParseJSON([]byte(value), "value.tfvars.json")
		} else {
			file, diags = parser.ParseHCL([]byte(value), "value.tfvars")
		}
		if diags.HasErrors() {
			return fmt.Errorf("invalid HCL: %s", diags.Error())
		}
		// tfvars are a flat set of attribute assignments; reject blocks and
		// other non-tfvars HCL so a bad value is caught here, not at deploy.
		if _, diags := file.Body.JustAttributes(); diags.HasErrors() {
			return fmt.Errorf("invalid tfvars: %s", diags.Error())
		}
		return nil
	default:
		return nil
	}
}
