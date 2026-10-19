package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const validCreateUIDefinition = `{
  "$schema": "https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#",
  "handler": "Microsoft.Azure.CreateUIDef",
  "version": "0.1.2-preview",
  "parameters": {
    "basics": [{"name": "vnetCIDR", "type": "Microsoft.Common.TextBox", "label": "Network CIDR"}],
    "steps": []
  }
}`

func TestValidateAzureUIDefinition(t *testing.T) {
	require.NoError(t, ValidateAzureUIDefinition("azure-bicep", ""))
	require.NoError(t, ValidateAzureUIDefinition("azure-bicep", "./arm/ui.json"))
	require.NoError(t, ValidateAzureUIDefinition("aws-cloudformation", ""))

	err := ValidateAzureUIDefinition("aws-cloudformation", "./arm/ui.json")
	require.ErrorContains(t, err, "only supported when type is azure-bicep")
}

func TestValidateCreateUIDefinition(t *testing.T) {
	require.NoError(t, ValidateCreateUIDefinition(""))
	require.NoError(t, ValidateCreateUIDefinition(validCreateUIDefinition))

	err := ValidateCreateUIDefinition(`{"handler":"Microsoft.Azure.CreateUIDef"}`)
	require.ErrorContains(t, err, "createUiDefinition")
}

func TestStackConfig_Parse_UIDefinition(t *testing.T) {
	aws := &StackConfig{
		Type:                    "aws-cloudformation",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		UIDefinition:            "./arm/ui.json",
		UIDefinitionContents:    validCreateUIDefinition,
	}
	require.ErrorContains(t, aws.parse(), "only supported when type is azure-bicep")

	azure := &StackConfig{
		Type:         "azure-bicep",
		UIDefinition: "./arm/ui.json",
	}
	require.NoError(t, azure.parse())
}

func TestStackConfig_UIDefinitionContentsRoundTrip(t *testing.T) {
	stack := &StackConfig{
		Type:                 "azure-bicep",
		UIDefinition:         "./arm/ui.json",
		UIDefinitionContents: validCreateUIDefinition,
	}
	body, err := json.Marshal(stack)
	require.NoError(t, err)

	var decoded StackConfig
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, validCreateUIDefinition, decoded.UIDefinitionContents)
	require.Equal(t, "./arm/ui.json", decoded.UIDefinition)
}
