package config

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	yaml "gopkg.in/yaml.v3"
)

const (
	InputTypeYAML = "yaml"
	InputTypeHCL  = "hcl"
)

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
		if _, diags := file.Body.JustAttributes(); diags.HasErrors() {
			return fmt.Errorf("invalid tfvars: %s", diags.Error())
		}
		return nil
	default:
		return nil
	}
}
