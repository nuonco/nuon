package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidateCustomNestedStackOutputs_NilStack(t *testing.T) {
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_NoAdditionalStacks(t *testing.T) {
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{},
	})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_NoCollisions(t *testing.T) {
	templateA := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  OutputA:
    Value: a
`
	templateB := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  OutputB:
    Value: b
`
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateA))
	})
	mux.HandleFunc("/b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateB))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "stack_a", TemplateURL: server.URL + "/a.yaml", Index: 0},
				{Name: "stack_b", TemplateURL: server.URL + "/b.yaml", Index: 1},
			},
		},
	})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_CollisionBetweenAdditionalStacks(t *testing.T) {
	templateA := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  SharedOutput:
    Value: a
`
	templateB := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  SharedOutput:
    Value: b
`
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateA))
	})
	mux.HandleFunc("/b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateB))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "stack_a", TemplateURL: server.URL + "/a.yaml", Index: 0},
				{Name: "stack_b", TemplateURL: server.URL + "/b.yaml", Index: 1},
			},
		},
	})
	require.Error(t, err)
	assert.True(t, config.IsWarningErr(err))
	assert.Contains(t, err.Error(), "SharedOutput")
	assert.Contains(t, err.Error(), "stack_a")
	assert.Contains(t, err.Error(), "stack_b")
}

func TestValidateCustomNestedStackOutputs_CollisionWithFirstClass(t *testing.T) {
	vpcTemplate := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  VPC:
    Value: vpc-id
  RunnerSubnet:
    Value: subnet-id
`
	additionalTemplate := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  VPC:
    Value: colliding-vpc
`
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(vpcTemplate))
	})
	mux.HandleFunc("/additional.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(additionalTemplate))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			VPCNestedTemplateURL: server.URL + "/vpc.yaml",
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "my_stack", TemplateURL: server.URL + "/additional.yaml", Index: 0},
			},
		},
	})
	require.Error(t, err)
	assert.True(t, config.IsWarningErr(err))
	assert.Contains(t, err.Error(), "VPC")
	assert.Contains(t, err.Error(), "first-class output will take precedence")
}

func TestValidateCustomNestedStackOutputs_UnfetchableTemplateSkipped(t *testing.T) {
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "my_stack", TemplateURL: "http://invalid.localhost.test/stack.yaml", Index: 0},
			},
		},
	})
	assert.NoError(t, err)
}
