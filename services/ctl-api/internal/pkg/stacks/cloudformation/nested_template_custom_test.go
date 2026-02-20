package cloudformation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cfn "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const mockAdditionalTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom nested stack for k8s namespace setup
Parameters:
  NuonInstallID:
    Description: The Nuon Install ID.
    Type: String
  NuonOrgID:
    Description: The Nuon Org ID.
    Type: String
  NuonAppID:
    Description: The Nuon App ID.
    Type: String
  NamespaceName:
    Description: The Kubernetes namespace to create.
    Type: String
    Default: app-namespace
  ServiceAccountName:
    Description: Service account name for the namespace.
    Type: String
Resources:
  Namespace:
    Type: Custom::KubernetesNamespace
    Properties:
      Name: !Ref NamespaceName
`

const mockAdditionalTemplate2YAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom nested stack for EKS access entries
Parameters:
  NuonInstallID:
    Description: The Nuon Install ID.
    Type: String
  AccessPrincipalArn:
    Description: The IAM principal ARN for EKS access.
    Type: String
  AccessType:
    Description: The type of EKS access entry.
    Type: String
    Default: namespace
Resources:
  AccessEntry:
    Type: Custom::EKSAccessEntry
    Properties:
      PrincipalArn: !Ref AccessPrincipalArn
`

func newTestInput(serverURL string, customStacks []config.CustomNestedStack) *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: serverURL + "/vpc.yaml",
				CustomNestedStacks:   customStacks,
			},
		},
	}
}

func TestGetCustomNestedStacks_Empty(t *testing.T) {
	tpl := &Templates{}
	inp := newTestInput("http://localhost", nil)
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)
	assert.Empty(t, result.resources)
	assert.Empty(t, result.params)
	assert.Nil(t, result.paramGroups)
}

func TestGetCustomNestedStacks_SingleStack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockAdditionalTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "k8s_namespaces", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	assert.Len(t, result.resources, 1)
	assert.Contains(t, result.resources, "K8SNamespaces")

	stack := result.resources["K8SNamespaces"]
	assert.Equal(t, []string{"VPC", "RunnerAutoScalingGroup"}, stack.AWSCloudFormationDependsOn)
	assert.Equal(t, "test-install-id", stack.Parameters["NuonInstallID"])
	assert.Equal(t, "test-app-id", stack.Parameters["NuonAppID"])
	assert.Equal(t, "test-org-id", stack.Parameters["NuonOrgID"])

	assert.Contains(t, result.params, "NamespaceName")
	assert.Contains(t, result.params, "ServiceAccountName")
	assert.NotContains(t, result.params, "NuonInstallID")
	assert.NotContains(t, result.params, "NuonOrgID")
	assert.NotContains(t, result.params, "NuonAppID")

	require.Len(t, result.paramGroups, 1)
	label := result.paramGroups[0]["Label"].(map[string]any)
	assert.Equal(t, "k8s_namespaces", label["default"])
}

func TestGetCustomNestedStacks_MultipleStacks_DependsOnChaining(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/template1.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateYAML))
	})
	mux.HandleFunc("/template2.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplate2YAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "k8s_namespaces", TemplateURL: server.URL + "/template1.yaml", Index: 0},
		{Name: "eks_access", TemplateURL: server.URL + "/template2.yaml", Index: 1},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	assert.Len(t, result.resources, 2)
	assert.Contains(t, result.resources, "K8SNamespaces")
	assert.Contains(t, result.resources, "EksAccess")

	assert.Equal(t, []string{"VPC", "RunnerAutoScalingGroup"}, result.resources["K8SNamespaces"].AWSCloudFormationDependsOn)
	assert.Equal(t, []string{"K8SNamespaces"}, result.resources["EksAccess"].AWSCloudFormationDependsOn)

	assert.Contains(t, result.params, "NamespaceName")
	assert.Contains(t, result.params, "ServiceAccountName")
	assert.Contains(t, result.params, "AccessPrincipalArn")
	assert.Contains(t, result.params, "AccessType")

	assert.Len(t, result.paramGroups, 2)
}

func TestGetCustomNestedStacks_MissingName(t *testing.T) {
	tpl := &Templates{}
	inp := newTestInput("http://localhost", []config.CustomNestedStack{
		{Name: "", TemplateURL: "http://example.com/stack.yaml", Index: 0},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestGetCustomNestedStacks_MissingTemplateURL(t *testing.T) {
	tpl := &Templates{}
	inp := newTestInput("http://localhost", []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: "", Index: 0},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template_url is required")
}

func TestGetCustomNestedStacks_ConflictWithExistingResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "vpc", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"Vpc": true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with existing resource")
}

func TestGetCustomNestedStacks_DuplicateNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 1},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate logical ID")
}

func TestGetCustomNestedStacks_ParameterConflictBetweenStacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "stack_a", TemplateURL: server.URL + "/stack.yaml", Index: 0},
		{Name: "stack_b", TemplateURL: server.URL + "/stack.yaml", Index: 1},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter")
	assert.Contains(t, err.Error(), "conflicts with stack")
}

func TestGetCustomNestedStacks_IndexDeterminesOrder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/template1.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateYAML))
	})
	mux.HandleFunc("/template2.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplate2YAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	// provide stacks in reverse index order — index 10 first, index 5 second
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "eks_access", TemplateURL: server.URL + "/template2.yaml", Index: 10},
		{Name: "k8s_namespaces", TemplateURL: server.URL + "/template1.yaml", Index: 5},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	// k8s_namespaces (index 5) should be first → depends on VPC/Runner
	assert.Equal(t, []string{"VPC", "RunnerAutoScalingGroup"}, result.resources["K8SNamespaces"].AWSCloudFormationDependsOn)
	// eks_access (index 10) should be second → depends on K8SNamespaces
	assert.Equal(t, []string{"K8SNamespaces"}, result.resources["EksAccess"].AWSCloudFormationDependsOn)
}

func TestGetCustomNestedStacks_DuplicateIndex(t *testing.T) {
	tpl := &Templates{}
	inp := newTestInput("http://localhost", []config.CustomNestedStack{
		{Name: "stack_a", TemplateURL: "http://example.com/a.yaml", Index: 1},
		{Name: "stack_b", TemplateURL: "http://example.com/b.yaml", Index: 1},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	_, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate index 1")
}

const mockVPCTemplateWithOutputsYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: VPC template with outputs
Parameters:
  NuonInstallID:
    Type: String
  ClusterName:
    Type: String
Outputs:
  VPC:
    Value: !Ref VPCResource
  RunnerSubnet:
    Value: !GetAtt VPCResource.Subnet
  PublicSubnets:
    Value: comma-separated-public-subnets
  PrivateSubnets:
    Value: comma-separated-private-subnets
Resources:
  VPCResource:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
`

const mockRunnerTemplateWithOutputsYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Runner ASG template with outputs
Parameters:
  SubnetId:
    Type: String
Outputs:
  RunnerInstanceRole:
    Value: !GetAtt RunnerRole.Arn
  RunnerInstanceRoleARN:
    Value: !GetAtt RunnerRole.Arn
Resources:
  RunnerRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: runner-role
`

const mockAdditionalTemplateWithVPCParamsYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom nested stack that needs VPC outputs
Parameters:
  NuonInstallID:
    Type: String
  VPC:
    Description: The VPC ID from the VPC stack
    Type: String
  RunnerSubnet:
    Description: The runner subnet from the VPC stack
    Type: String
  CustomParam:
    Description: A custom parameter not from any first-class stack
    Type: String
    Default: my-default
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      VpcId: !Ref VPC
`

const mockAdditionalTemplateWithNamespacesYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  Namespaces:
    Description: Comma-separated list of namespaces
    Type: String
  CustomParam:
    Description: A custom parameter
    Type: String
    Default: default-value
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      Namespaces: !Ref Namespaces
`

const mockAdditionalTemplateNamespacesOnlyYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  Namespaces:
    Description: Comma-separated list of namespaces
    Type: String
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      Namespaces: !Ref Namespaces
`

func TestGetCustomNestedStacks_ExplicitParameterMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateWithNamespacesYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	nsVal := "sourdough,persimmon"
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{
			Name:        "my_stack",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "{{.nuon.install.inputs.namespaces}}",
			},
		},
	})
	inp.Install.CurrentInstallInputs = &app.InstallInputs{
		Values: pgtype.Hstore{"namespaces": &nsVal},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, "sourdough,persimmon", stack.Parameters["Namespaces"])
	assert.NotContains(t, result.params, "Namespaces")
	assert.Contains(t, result.params, "CustomParam")

	require.Len(t, result.paramGroups, 1)
	params := result.paramGroups[0]["Parameters"].([]string)
	assert.Contains(t, params, "CustomParam")
	assert.NotContains(t, params, "Namespaces")
}

func TestGetCustomNestedStacks_ExplicitParameterMappingAcrossMultipleStacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateNamespacesOnlyYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	nsVal := "sourdough,persimmon"
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{
			Name:        "stack_a",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "{{.nuon.install.inputs.namespaces}}",
			},
		},
		{
			Name:        "stack_b",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       1,
			Parameters: map[string]string{
				"Namespaces": "{{.nuon.install.inputs.namespaces}}",
			},
		},
	})
	inp.Install.CurrentInstallInputs = &app.InstallInputs{
		Values: pgtype.Hstore{"namespaces": &nsVal},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	assert.Len(t, result.resources, 2)
	assert.Equal(t, "sourdough,persimmon", result.resources["StackA"].Parameters["Namespaces"])
	assert.Equal(t, "sourdough,persimmon", result.resources["StackB"].Parameters["Namespaces"])
	assert.NotContains(t, result.params, "Namespaces")
	assert.Empty(t, result.paramGroups)
}

func TestGetCustomNestedStacks_ExplicitParameterEmptyWhenNoInstallInputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateNamespacesOnlyYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{
			Name:        "my_stack",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "{{.nuon.install.inputs.namespaces}}",
			},
		},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, "", stack.Parameters["Namespaces"])
	assert.NotContains(t, result.params, "Namespaces")
}

func TestGetCustomNestedStacks_FirstClassOutputWiring(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockVPCTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/runner.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockRunnerTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/additional.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalTemplateWithVPCParamsYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
				RunnerNestedTemplateURL: server.URL + "/runner.yaml",
				CustomNestedStacks: []config.CustomNestedStack{
					{Name: "my_stack", TemplateURL: server.URL + "/additional.yaml", Index: 0},
				},
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, cfn.GetAtt("VPC", "Outputs.VPC"), stack.Parameters["VPC"])
	assert.Equal(t, cfn.GetAtt("VPC", "Outputs.RunnerSubnet"), stack.Parameters["RunnerSubnet"])

	assert.NotContains(t, result.params, "VPC", "VPC should not be hoisted as a top-level parameter")
	assert.NotContains(t, result.params, "RunnerSubnet", "RunnerSubnet should not be hoisted as a top-level parameter")

	assert.Contains(t, result.params, "CustomParam", "CustomParam should still be hoisted")

	require.Len(t, result.paramGroups, 1)
	params := result.paramGroups[0]["Parameters"].([]string)
	assert.Contains(t, params, "CustomParam")
	assert.NotContains(t, params, "VPC")
	assert.NotContains(t, params, "RunnerSubnet")
}

func TestGetCustomNestedStacks_RunnerOutputWiring(t *testing.T) {
	mockAdditionalWithRunnerParams := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  RunnerInstanceRole:
    Description: The runner IAM role
    Type: String
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      RoleArn: !Ref RunnerInstanceRole
`
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockVPCTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/runner.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockRunnerTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/additional.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAdditionalWithRunnerParams))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
				RunnerNestedTemplateURL: server.URL + "/runner.yaml",
				CustomNestedStacks: []config.CustomNestedStack{
					{Name: "my_stack", TemplateURL: server.URL + "/additional.yaml", Index: 0},
				},
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, cfn.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"), stack.Parameters["RunnerInstanceRole"])
	assert.NotContains(t, result.params, "RunnerInstanceRole")
}

func TestGetCustomNestedStacks_ExplicitClusterNameMapping(t *testing.T) {
	mockTemplateWithClusterName := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  ClusterName:
    Type: String
  NuonInstallID:
    Type: String
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      ClusterName: !Ref ClusterName
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateWithClusterName))
	}))
	defer server.Close()

	tpl := &Templates{}
	clusterVal := "my-cluster"
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{
			Name:        "sg_access",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"ClusterName": "{{.nuon.install.inputs.cluster_name}}",
			},
		},
	})
	inp.Install.CurrentInstallInputs = &app.InstallInputs{
		Values: pgtype.Hstore{"cluster_name": &clusterVal},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["SgAccess"]
	require.NotNil(t, stack)

	assert.Equal(t, "my-cluster", stack.Parameters["ClusterName"])
	assert.Equal(t, "test-install-id", stack.Parameters["NuonInstallID"])
	assert.NotContains(t, result.params, "ClusterName")
	assert.NotContains(t, result.params, "NuonInstallID")
}

func TestGetCustomNestedStacks_ReservedParamsNotInjectedWhenAbsent(t *testing.T) {
	mockTemplateNoReservedParams := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  CustomParam:
    Type: String
    Default: default-value
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      Param: !Ref CustomParam
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateNoReservedParams))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	_, hasNuonInstallID := stack.Parameters["NuonInstallID"]
	_, hasNuonAppID := stack.Parameters["NuonAppID"]
	_, hasNuonOrgID := stack.Parameters["NuonOrgID"]
	assert.False(t, hasNuonInstallID, "NuonInstallID should not be injected when template doesn't declare it")
	assert.False(t, hasNuonAppID, "NuonAppID should not be injected when template doesn't declare it")
	assert.False(t, hasNuonOrgID, "NuonOrgID should not be injected when template doesn't declare it")

	assert.Contains(t, result.params, "CustomParam")
}

func TestGetCustomNestedStacks_RoleParamInjectedWhenDeclared(t *testing.T) {
	mockTemplateWithRoleParam := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  EnableRunnerProvision:
    Description: Enable the provision role
    Type: String
    Default: "true"
    AllowedValues:
      - "true"
      - "false"
  CustomParam:
    Type: String
    Default: default-value
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      Param: !Ref CustomParam
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateWithRoleParam))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	inp.AppCfg.PermissionsConfig = app.AppPermissionsConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				Type:                         "runner_provision",
				CloudFormationStackName:      "RunnerProvision",
				CloudFormationStackParamName: "EnableRunnerProvision",
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, cfn.Ref("EnableRunnerProvision"), stack.Parameters["EnableRunnerProvision"])
	assert.Equal(t, "test-install-id", stack.Parameters["NuonInstallID"])
	assert.NotContains(t, result.params, "EnableRunnerProvision", "role param should not be hoisted")
	assert.NotContains(t, result.params, "NuonInstallID")
	assert.Contains(t, result.params, "CustomParam")

	// Role resources are conditional, so they must NOT appear in DependsOn
	// (otherwise CloudFormation fails with "Unresolved resource dependencies"
	// when the role condition is false).
	assert.NotContains(t, stack.AWSCloudFormationDependsOn, "RunnerProvision")
}

func TestGetCustomNestedStacks_RoleParamNotInjectedWhenAbsent(t *testing.T) {
	mockTemplateNoRoleParam := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  CustomParam:
    Type: String
    Default: default-value
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      Param: !Ref CustomParam
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateNoRoleParam))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	inp.AppCfg.PermissionsConfig = app.AppPermissionsConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				Type:                         "runner_provision",
				CloudFormationStackName:      "RunnerProvision",
				CloudFormationStackParamName: "EnableRunnerProvision",
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	_, hasRoleParam := stack.Parameters["EnableRunnerProvision"]
	assert.False(t, hasRoleParam, "role param should not be injected when template doesn't declare it")
	assert.NotContains(t, stack.AWSCloudFormationDependsOn, "RunnerProvision")
	assert.Equal(t, []string{"VPC", "RunnerAutoScalingGroup"}, stack.AWSCloudFormationDependsOn)
}

func TestGetCustomNestedStacks_MultipleRolesPartialMatch(t *testing.T) {
	mockTemplateOneRole := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  EnableRunnerProvision:
    Type: String
    Default: "true"
Resources:
  MyResource:
    Type: Custom::Resource
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateOneRole))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := newTestInput(server.URL, []config.CustomNestedStack{
		{Name: "my_stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	inp.AppCfg.PermissionsConfig = app.AppPermissionsConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				Type:                         "runner_provision",
				CloudFormationStackName:      "RunnerProvision",
				CloudFormationStackParamName: "EnableRunnerProvision",
			},
			{
				Type:                         "runner_deprovision",
				CloudFormationStackName:      "RunnerDeprovision",
				CloudFormationStackParamName: "EnableRunnerDeprovision",
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stack := result.resources["MyStack"]
	require.NotNil(t, stack)

	assert.Equal(t, cfn.Ref("EnableRunnerProvision"), stack.Parameters["EnableRunnerProvision"])
	_, hasDeprov := stack.Parameters["EnableRunnerDeprovision"]
	assert.False(t, hasDeprov, "deprovision param should not be injected when template doesn't declare it")

	// Conditional role resources must NOT be in DependsOn.
	assert.NotContains(t, stack.AWSCloudFormationDependsOn, "RunnerProvision")
	assert.NotContains(t, stack.AWSCloudFormationDependsOn, "RunnerDeprovision")
}

const mockTemplateWithOutputsA = `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  CustomParamA:
    Type: String
    Default: default-a
Resources:
  MyResource:
    Type: Custom::Resource
Outputs:
  SharedSubnetID:
    Value: !Ref MyResource
  StackAOnlyOutput:
    Value: some-value
`

const mockTemplateConsumingOutputs = `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  SharedSubnetID:
    Description: Subnet from the previous stack
    Type: String
  StackAOnlyOutput:
    Description: Output only from stack A
    Type: String
  CustomParamB:
    Type: String
    Default: default-b
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      SubnetId: !Ref SharedSubnetID
`

const mockTemplateThirdStack = `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  StackAOnlyOutput:
    Description: Output from stack A, passed through
    Type: String
  CustomParamC:
    Type: String
    Default: default-c
Resources:
  MyResource:
    Type: Custom::Resource
Outputs:
  FinalOutput:
    Value: final
`

func TestGetCustomNestedStacks_InterStackOutputWiring(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockVPCTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/stack_a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateWithOutputsA))
	})
	mux.HandleFunc("/stack_b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateConsumingOutputs))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: server.URL + "/vpc.yaml",
				CustomNestedStacks: []config.CustomNestedStack{
					{Name: "stack_a", TemplateURL: server.URL + "/stack_a.yaml", Index: 0},
					{Name: "stack_b", TemplateURL: server.URL + "/stack_b.yaml", Index: 1},
				},
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stackB := result.resources["StackB"]
	require.NotNil(t, stackB)

	assert.Equal(t, cfn.GetAtt("StackA", "Outputs.SharedSubnetID"), stackB.Parameters["SharedSubnetID"])
	assert.Equal(t, cfn.GetAtt("StackA", "Outputs.StackAOnlyOutput"), stackB.Parameters["StackAOnlyOutput"])

	assert.NotContains(t, result.params, "SharedSubnetID")
	assert.NotContains(t, result.params, "StackAOnlyOutput")
	assert.Contains(t, result.params, "CustomParamB")

	stackA := result.resources["StackA"]
	require.NotNil(t, stackA)
	assert.Contains(t, result.params, "CustomParamA")
}

func TestGetCustomNestedStacks_InterStackOutputWiring_ThreeStackChain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockVPCTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/stack_a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateWithOutputsA))
	})
	mux.HandleFunc("/stack_b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateConsumingOutputs))
	})
	mux.HandleFunc("/stack_c.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateThirdStack))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: server.URL + "/vpc.yaml",
				CustomNestedStacks: []config.CustomNestedStack{
					{Name: "stack_a", TemplateURL: server.URL + "/stack_a.yaml", Index: 0},
					{Name: "stack_b", TemplateURL: server.URL + "/stack_b.yaml", Index: 1},
					{Name: "stack_c", TemplateURL: server.URL + "/stack_c.yaml", Index: 2},
				},
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	assert.Len(t, result.resources, 3)

	stackC := result.resources["StackC"]
	require.NotNil(t, stackC)

	assert.Equal(t, cfn.GetAtt("StackA", "Outputs.StackAOnlyOutput"), stackC.Parameters["StackAOnlyOutput"])
	assert.NotContains(t, result.params, "StackAOnlyOutput")
	assert.Contains(t, result.params, "CustomParamC")
}

func TestGetCustomNestedStacks_FirstClassOutputWinsOverInterStack(t *testing.T) {
	mockTemplateOutputMatchingVPC := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
Resources:
  MyResource:
    Type: Custom::Resource
Outputs:
  VPC:
    Value: additional-stack-vpc-value
  RunnerSubnet:
    Value: additional-stack-subnet-value
`
	mockTemplateConsumingVPCParam := `
AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonInstallID:
    Type: String
  VPC:
    Description: Should come from first-class VPC stack, not from stack_a
    Type: String
  RunnerSubnet:
    Description: Should come from first-class VPC stack, not from stack_a
    Type: String
Resources:
  MyResource:
    Type: Custom::Resource
    Properties:
      VpcId: !Ref VPC
`
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockVPCTemplateWithOutputsYAML))
	})
	mux.HandleFunc("/stack_a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateOutputMatchingVPC))
	})
	mux.HandleFunc("/stack_b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockTemplateConsumingVPCParam))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: server.URL + "/vpc.yaml",
				CustomNestedStacks: []config.CustomNestedStack{
					{Name: "stack_a", TemplateURL: server.URL + "/stack_a.yaml", Index: 0},
					{Name: "stack_b", TemplateURL: server.URL + "/stack_b.yaml", Index: 1},
				},
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{"VPC": true, "RunnerAutoScalingGroup": true})
	require.NoError(t, err)

	stackB := result.resources["StackB"]
	require.NotNil(t, stackB)

	assert.Equal(t, cfn.GetAtt("VPC", "Outputs.VPC"), stackB.Parameters["VPC"],
		"VPC param should be wired from first-class VPC stack, not from stack_a")
	assert.Equal(t, cfn.GetAtt("VPC", "Outputs.RunnerSubnet"), stackB.Parameters["RunnerSubnet"],
		"RunnerSubnet param should be wired from first-class VPC stack, not from stack_a")
}

func TestSanitizeLogicalID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"k8s_namespaces", "K8SNamespaces"},
		{"eks-access-entries", "EksAccessEntries"},
		{"simple", "Simple"},
		{"my stack name", "MyStackName"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, sanitizeLogicalID(tc.input))
		})
	}
}
