package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const armTemplateJSON = `{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "resources": []
}`

const bicepSource = `param location string = resourceGroup().location

resource sa 'Microsoft.Storage/storageAccounts@2022-09-01' = {
  name: 'acmestorage'
  location: location
}`

func TestValidateAzureCustomNestedStacks_RejectsBicepTemplateURL(t *testing.T) {
	err := ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "networking", Index: 0, TemplateURL: "./arm/networking.json", Contents: armTemplateJSON},
		{Name: "storage", Index: 1, TemplateURL: "./arm/storage.bicep", Contents: armTemplateJSON},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom_nested_stacks[1] (storage)")
	assert.Contains(t, err.Error(), `template_url "./arm/storage.bicep" is Bicep source`)
	assert.Contains(t, err.Error(), "az bicep build --file ./arm/storage.bicep --outfile ./arm/storage.json")
}

func TestValidateAzureCustomNestedStacks_RejectsBicepTemplateURLUppercase(t *testing.T) {
	err := ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "storage", Index: 0, TemplateURL: "./arm/Storage.BICEP"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "az bicep build --file ./arm/Storage.BICEP --outfile ./arm/Storage.json")
}

func TestValidateAzureCustomNestedStacks_RejectsNonARMJSONContents(t *testing.T) {
	err := ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "storage", Index: 0, TemplateURL: "./arm/storage.json", Contents: bicepSource},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom_nested_stacks[0] (storage)")
	assert.Contains(t, err.Error(), `contents of template_url "./arm/storage.json" are not valid ARM JSON`)
	assert.Contains(t, err.Error(), "az bicep build --file ./arm/storage.bicep --outfile ./arm/storage.json")
}

func TestValidateAzureCustomNestedStacks_RejectsNonObjectContents(t *testing.T) {
	err := ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "storage", Index: 0, TemplateURL: "https://example.com/storage.json", Contents: `["not", "a", "template"]`},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "are not valid ARM JSON")
}

func TestValidateAzureCustomNestedStacks_RejectsJSONWithoutARMContract(t *testing.T) {
	err := ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "storage", Index: 0, TemplateURL: "./arm/storage.json", Contents: `{"resources":[]}`},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "are not valid ARM JSON")
}

func TestValidateAzureCustomNestedStacks_AllowsValidARMJSON(t *testing.T) {
	require.NoError(t, ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "networking", Index: 0, TemplateURL: "./arm/networking.json", Contents: armTemplateJSON},
		{Name: "storage", Index: 1, TemplateURL: "https://example.com/storage.json", Contents: armTemplateJSON},
	}))
}

func TestValidateAzureCustomNestedStacks_SkipsUnfetchedContents(t *testing.T) {
	require.NoError(t, ValidateAzureCustomNestedStacks("azure-bicep", []CustomNestedStack{
		{Name: "storage", Index: 0, TemplateURL: "https://example.com/storage.json"},
	}))
}

func TestValidateAzureCustomNestedStacks_SkipsNonAzureStackTypes(t *testing.T) {
	stacks := []CustomNestedStack{
		{Name: "namespaces", Index: 0, TemplateURL: "./cfn/namespaces.yaml", Contents: "Resources: {}"},
		{Name: "bicep_named", Index: 1, TemplateURL: "./cfn/thing.bicep", Contents: "Resources: {}"},
	}

	for _, stackType := range []string{"aws-cloudformation", "gcp-terraform", ""} {
		require.NoError(t, ValidateAzureCustomNestedStacks(stackType, stacks), stackType)
	}
}

func TestValidateAzureCustomNestedStacks_NoStacks(t *testing.T) {
	require.NoError(t, ValidateAzureCustomNestedStacks("azure-bicep", nil))
}
