package cloudformation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const mockContractParamTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack that binds to the frozen VPC/runner contract.
Parameters:
  NuonInstallID:
    Type: String
  VPC:
    Description: The VPC id.
    Type: String
  CIDRBlock:
    Type: String
  RunnerSubnet:
    Type: String
  PublicSubnets:
    Type: String
  PrivateSubnets:
    Type: String
  ExtraSetting:
    Type: String
    Default: fallback-value
Resources:
  Namespace:
    Type: Custom::KubernetesNamespace
    Properties:
      VpcId: !Ref VPC
Outputs:
  NamespaceName:
    Value: !Ref Namespace
`

const mockUnboundParamTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack with a parameter that cannot be resolved in CustomStacksOnly mode.
Parameters:
  Unresolvable:
    Type: String
Resources:
  Namespace:
    Type: Custom::KubernetesNamespace
    Properties:
      Setting: !Ref Unresolvable
`

const mockHoistableParamTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack with a simple install-input-reference parameter and a complex-expression parameter.
Parameters:
  Namespaces:
    Type: String
  RootDomain:
    Type: String
Resources:
  Namespace:
    Type: Custom::KubernetesNamespace
    Properties:
      Namespaces: !Ref Namespaces
      RootDomain: !Ref RootDomain
`

const mockAParamTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack whose hoisted parameter name is misattributed to another configured stack under the flat encoding.
Parameters:
  breplicaArn:
    Type: String
Resources:
  Thing:
    Type: Custom::Thing
    Properties:
      Arn: !Ref breplicaArn
`

const mockATemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack whose output is misattributed to another configured stack under the flat encoding.
Resources:
  Thing:
    Type: Custom::Thing
Outputs:
  breplicaArn:
    Value: !GetAtt Thing.Arn
`

const mockABReplicaTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack with no outputs of its own, whose logical ID nonetheless shadows another stack's output name.
Resources:
  Thing:
    Type: Custom::Thing
`

const mockInstallInputNameMatchTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack whose parameter name matches an install input's CloudFormationStackParamName, with no template-supplied default.
Parameters:
  InstallRootDomain:
    Type: String
Resources:
  Namespace:
    Type: Custom::KubernetesNamespace
    Properties:
      RootDomain: !Ref InstallRootDomain
`

const mockProducerTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack producing an output another custom stack consumes.
Resources:
  Bucket:
    Type: Custom::Bucket
Outputs:
  BucketArn:
    Value: !GetAtt Bucket.Arn
`

const mockConsumerTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Custom stack consuming another custom stack's output by name match.
Parameters:
  BucketArn:
    Type: String
Resources:
  Policy:
    Type: Custom::Policy
    Properties:
      BucketArn: !Ref BucketArn
`

func newCustomStacksOnlyInput(customStacks []config.CustomNestedStack) *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				CustomNestedStacks: customStacks,
			},
			InputConfig: app.AppInputConfig{
				AppInputs: []app.AppInput{
					{Name: "namespaces", Source: app.AppInputSourceCustomer},
					{Name: "some_arn", Source: app.AppInputSourceCustomer},
				},
			},
		},
		Settings:         &app.RunnerGroupSettings{},
		CustomStacksOnly: true,
	}
}

func TestGetAWSCustomStacksOnlyTemplate_NoQuicklinkResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockContractParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "k8s-namespaces", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)

	for _, forbidden := range []string{
		"VPC", "RunnerSecurityGroup", "RunnerAutoScalingGroup",
		"RunnerCloudWatchLogGroup", "RunnerCloudWatchLogStream", "RunnerCloudWatchLogPolicy",
		"RunnerPhoneHome", "RunnerPhoneHomeRole", "PhoneHomeProps",
	} {
		assert.NotContains(t, tmpl.Resources, forbidden)
	}
	// no IAM roles, no secrets, no telemetry resources of any kind
	assert.Len(t, tmpl.Resources, 1)
	assert.Contains(t, tmpl.Resources, "K8SNamespaces")
}

func TestGetAWSCustomStacksOnlyTemplate_ContractParamsAlwaysDeclared(t *testing.T) {
	t.Run("with zero custom stacks referencing them", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(mockProducerTemplateYAML))
		}))
		defer server.Close()

		tpl := &Templates{cfg: &internal.Config{}}
		inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
			{Name: "producer", TemplateURL: server.URL + "/stack.yaml", Index: 0},
		})

		tmpl, err := tpl.getAWSTemplate(inp)
		require.NoError(t, err)

		for _, name := range AWSCustomStacksOnlyContractParams {
			require.Contains(t, tmpl.Parameters, name)
			assert.Equal(t, "String", tmpl.Parameters[name].Type)
		}
	})

	t.Run("no custom stacks at all is a hard error", func(t *testing.T) {
		tpl := &Templates{cfg: &internal.Config{}}
		inp := newCustomStacksOnlyInput(nil)

		_, err := tpl.getAWSTemplate(inp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no custom_nested_stacks configured")
	})
}

func TestGetAWSCustomStacksOnlyTemplate_ContractParamValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockContractParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "k8s-namespaces", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	stack := result.resources["K8SNamespaces"]
	require.NotNil(t, stack)

	assert.Equal(t, cloudformation.Ref("VPC"), stack.Parameters["VPC"])
	assert.Equal(t, cloudformation.Ref("CIDRBlock"), stack.Parameters["CIDRBlock"])
	assert.Equal(t, cloudformation.Ref("RunnerSubnet"), stack.Parameters["RunnerSubnet"])
	assert.Equal(t, cloudformation.Ref("PublicSubnets"), stack.Parameters["PublicSubnets"])
	assert.Equal(t, cloudformation.Ref("PrivateSubnets"), stack.Parameters["PrivateSubnets"])

	// contract params should not additionally be exposed as top-level custom-stack
	// parameters (they're already declared as the frozen contract params)
	assert.NotContains(t, result.params, "VPC")
	assert.NotContains(t, result.params, "CIDRBlock")
	assert.NotContains(t, result.params, "RunnerSubnet")
	assert.NotContains(t, result.params, "PublicSubnets")
	assert.NotContains(t, result.params, "PrivateSubnets")

	// ExtraSetting has a template default, so it's still exposed for override
	assert.Contains(t, result.params, "ExtraSetting")
}

func TestGetAWSCustomStacksOnlyTemplate_CrossStackOutputBinding(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/producer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockProducerTemplateYAML))
	})
	mux.HandleFunc("/consumer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockConsumerTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "producer", TemplateURL: server.URL + "/producer.yaml", Index: 0},
		{Name: "consumer", TemplateURL: server.URL + "/consumer.yaml", Index: 1},
	})

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	consumer := result.resources["Consumer"]
	require.NotNil(t, consumer)
	assert.Equal(t, cloudformation.GetAtt("Producer", "Outputs.BucketArn"), consumer.Parameters["BucketArn"])
}

func TestGetAWSCustomStacksOnlyTemplate_DependsOnOrdering(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/producer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockProducerTemplateYAML))
	})
	mux.HandleFunc("/consumer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockConsumerTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "producer", TemplateURL: server.URL + "/producer.yaml", Index: 0},
		{Name: "consumer", TemplateURL: server.URL + "/consumer.yaml", Index: 1},
	})

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	assert.Empty(t, result.resources["Producer"].AWSCloudFormationDependsOn)
	assert.Equal(t, []string{"Producer"}, result.resources["Consumer"].AWSCloudFormationDependsOn)
}

func TestGetAWSCustomStacksOnlyTemplate_Outputs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/producer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockProducerTemplateYAML))
	})
	mux.HandleFunc("/contract.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockContractParamTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	stackNames := []string{"producer", "k8s-namespaces"}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: stackNames[0], TemplateURL: server.URL + "/producer.yaml", Index: 0},
		{Name: stackNames[1], TemplateURL: server.URL + "/contract.yaml", Index: 1},
	})

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)

	require.Contains(t, tmpl.Outputs, "ProducerBucketArn")
	assert.Equal(t, cloudformation.GetAtt("Producer", "Outputs.BucketArn"), tmpl.Outputs["ProducerBucketArn"].Value)

	require.Contains(t, tmpl.Outputs, "K8SNamespacesNamespaceName")
	assert.Equal(t, cloudformation.GetAtt("K8SNamespaces", "Outputs.NamespaceName"), tmpl.Outputs["K8SNamespacesNamespaceName"].Value)

	// round-trip: given the known stack names and the flat output map a deployed
	// stack would return, reconstruct the custom_nested_stacks phone-home shape.
	flat := map[string]string{
		"ProducerBucketArn":          "arn:aws:s3:::example-bucket",
		"K8SNamespacesNamespaceName": "app-namespace",
	}
	split := SplitCustomStacksOnlyOutputs(flat, stackNames)
	require.Contains(t, split, "producer")
	assert.Equal(t, "arn:aws:s3:::example-bucket", split["producer"]["outputs"]["BucketArn"])
	require.Contains(t, split, "k8s-namespaces")
	assert.Equal(t, "app-namespace", split["k8s-namespaces"]["outputs"]["NamespaceName"])
}

func TestGetAWSCustomStacksOnlyTemplate_AmbiguousOutputNameIsHardError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockATemplateYAML))
	})
	mux.HandleFunc("/abreplica.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockABReplicaTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "a", TemplateURL: server.URL + "/a.yaml", Index: 0},
		{Name: "abreplica", TemplateURL: server.URL + "/abreplica.yaml", Index: 1},
	})

	// Confirm by hand, against the real longest-prefix-match algorithm, that this
	// genuinely misattributes: stack "a" declares output "breplicaArn", flattening
	// to "AbreplicaArn". Stack "abreplica" declares no outputs of its own, but its
	// logical ID "Abreplica" (9 chars) is also a prefix of "AbreplicaArn" and is
	// longer than "A" (1 char), so SplitCustomStacksOnlyOutputs's longest-first
	// search matches "abreplica" instead of "a", handing "abreplica" a phantom
	// output key "Arn" instead of the real value landing on "a"/"breplicaArn".
	require.Equal(t, "A", sanitizeLogicalID("a"))
	require.Equal(t, "Abreplica", sanitizeLogicalID("abreplica"))
	split := SplitCustomStacksOnlyOutputs(
		map[string]string{"AbreplicaArn": "some-value"},
		[]string{"a", "abreplica"},
	)
	require.Equal(t, "some-value", split["abreplica"]["outputs"]["Arn"])
	require.NotContains(t, split, "a")

	_, err := tpl.getAWSTemplate(inp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AbreplicaArn")
	assert.Contains(t, err.Error(), `"a"`)
	assert.Contains(t, err.Error(), `"abreplica"`)
	assert.Contains(t, err.Error(), "breplicaArn")
	assert.Contains(t, err.Error(), "Arn")
	assert.Contains(t, err.Error(), "rename one of the two custom stacks")
}

func TestGetAWSCustomStacksOnlyTemplate_NonCollidingOutputsStillRender(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/producer.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockProducerTemplateYAML))
	})
	mux.HandleFunc("/contract.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockContractParamTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "producer", TemplateURL: server.URL + "/producer.yaml", Index: 0},
		{Name: "k8s-namespaces", TemplateURL: server.URL + "/contract.yaml", Index: 1},
	})

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)
	assert.Contains(t, tmpl.Outputs, "ProducerBucketArn")
	assert.Contains(t, tmpl.Outputs, "K8SNamespacesNamespaceName")
}

func TestGetAWSCustomStacksOnlyTemplate_UnbindableParameterIsHardError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockUnboundParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "broken-stack", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})

	_, err := tpl.getAWSTemplate(inp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken-stack")
	assert.Contains(t, err.Error(), "Unresolvable")
	assert.Contains(t, err.Error(), "VPC")
	assert.Contains(t, err.Error(), "CIDRBlock")
	assert.Contains(t, err.Error(), "RunnerSubnet")
	assert.Contains(t, err.Error(), "PublicSubnets")
	assert.Contains(t, err.Error(), "PrivateSubnets")
}

func TestGetAWSCustomStacksOnlyTemplate_SimpleInstallInputParameterIsHoisted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockHoistableParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{
			Name:        "k8s-namespaces",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "app-namespace",
				"RootDomain": "install123.example.com",
			},
		},
	})
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"k8s-namespaces": {
			"Namespaces": "{{.nuon.install.inputs.namespaces}}",
			"RootDomain": "install123.example.com",
		},
	}

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	stack := result.resources["K8SNamespaces"]
	require.NotNil(t, stack)
	assert.Equal(t, cloudformation.Ref("K8SNamespacesNamespaces"), stack.Parameters["Namespaces"])
	assert.Equal(t, "install123.example.com", stack.Parameters["RootDomain"])

	require.Contains(t, result.inputParameters, "k8s-namespaces")
	assert.Equal(t, map[string]string{"K8SNamespacesNamespaces": "namespaces"}, result.inputParameters["k8s-namespaces"])

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)
	require.Contains(t, tmpl.Parameters, "K8SNamespacesNamespaces")
	assert.Equal(t, "String", tmpl.Parameters["K8SNamespacesNamespaces"].Type)
	assert.NotContains(t, tmpl.Parameters, "K8SNamespacesRootDomain")
}

func TestGetAWSCustomStacksOnlyTemplate_VendorInputParameterIsBaked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockHoistableParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{
			Name:        "k8s-namespaces",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "vendor-namespace",
				"RootDomain": "example.com",
			},
		},
	})
	inp.AppCfg.InputConfig.AppInputs = []app.AppInput{
		{Name: "namespaces", Source: app.AppInputSourceVendor},
	}
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"k8s-namespaces": {"Namespaces": "{{.nuon.install.inputs.namespaces}}"},
	}

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)
	assert.NotContains(t, tmpl.Parameters, "K8SNamespacesNamespaces")

	stack := tmpl.Resources["K8SNamespaces"].(*nestedcloudformation.Stack)
	assert.Equal(t, "vendor-namespace", stack.Parameters["Namespaces"])
	require.Empty(t, ExtractAndStripCustomStacksInputParameters(tmpl))
}

func TestGetAWSCustomStacksOnlyTemplate_ComplexExpressionParameterIsBaked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockHoistableParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{
			Name:        "k8s-namespaces",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			Parameters: map[string]string{
				"Namespaces": "app-namespace",
				"RootDomain": "install123.example.com",
			},
		},
	})
	// RootDomain's unrendered form is a conditional template, not a whole-value
	// install-input reference, so ParseInstallInputReference must reject it and
	// the already-rendered literal in stack.Parameters must be baked verbatim,
	// exactly as before this feature existed.
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"k8s-namespaces": {
			"Namespaces": "app-namespace",
			"RootDomain": "{{ if .nuon.install.inputs.root_domain }}{{ .nuon.install.inputs.root_domain }}{{ else }}{{ .nuon.install.id }}.example.com{{ end }}",
		},
	}

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	stack := result.resources["K8SNamespaces"]
	require.NotNil(t, stack)
	assert.Equal(t, "install123.example.com", stack.Parameters["RootDomain"])
	assert.NotContains(t, result.inputParameters["k8s-namespaces"], "K8SNamespacesRootDomain")
}

func TestGetAWSCustomStacksOnlyTemplate_HoistedParameterNameCollisionIsHardError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockAParamTemplateYAML))
	})
	mux.HandleFunc("/abreplica.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockABReplicaTemplateYAML))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{
			Name:        "a",
			TemplateURL: server.URL + "/a.yaml",
			Index:       0,
			Parameters:  map[string]string{"breplicaArn": "some-value"},
		},
		{Name: "abreplica", TemplateURL: server.URL + "/abreplica.yaml", Index: 1},
	})
	// Stack "a" hoists parameter "breplicaArn" to top-level name "AbreplicaArn"
	// (sanitizeLogicalID("a") + "breplicaArn"). Stack "abreplica" declares no
	// parameters of its own, but its logical ID "Abreplica" is a longer prefix
	// match against "AbreplicaArn" than "a"'s own logical ID "A" -- the same
	// ambiguity class TestGetAWSCustomStacksOnlyTemplate_AmbiguousOutputNameIsHardError
	// exercises for outputs.
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"a": {"breplicaArn": "{{.nuon.install.inputs.some_arn}}"},
	}

	_, err := tpl.getAWSTemplate(inp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AbreplicaArn")
	assert.Contains(t, err.Error(), `"a"`)
	assert.Contains(t, err.Error(), `"abreplica"`)
	assert.Contains(t, err.Error(), "rename one of the two custom stacks")
}

func TestGetCustomNestedStacks_QuicklinkPathUnaffectedByHoistableParameter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockHoistableParamTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	stackCfg := config.CustomNestedStack{
		Name:        "k8s-namespaces",
		TemplateURL: server.URL + "/stack.yaml",
		Index:       0,
		Parameters: map[string]string{
			"Namespaces": "app-namespace",
			"RootDomain": "install123.example.com",
		},
	}

	// Same config, same unrendered install-input reference on Namespaces, as
	// TestGetAWSCustomStacksOnlyTemplate_SimpleInstallInputParameterIsHoisted --
	// the only difference below is CustomStacksOnly: false. The hoist branch
	// must be gated on that flag alone, so the quicklink path (CustomStacksOnly:
	// false) has to bake the literal exactly as it did before this feature
	// existed, with zero new top-level parameters.
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{stackCfg})
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"k8s-namespaces": {"Namespaces": "{{.nuon.install.inputs.namespaces}}"},
	}
	inp.CustomStacksOnly = false

	tb := tagBuilder{installID: inp.Install.ID}
	result, err := tpl.getCustomNestedStacks(inp, tb, map[string]bool{})
	require.NoError(t, err)

	stack := result.resources["K8SNamespaces"]
	require.NotNil(t, stack)
	assert.Equal(t, "app-namespace", stack.Parameters["Namespaces"])
	assert.NotContains(t, result.params, "K8SNamespacesNamespaces")
	assert.Empty(t, result.inputParameters)
}

func rootDomainInstallInput() app.AppInput {
	return app.AppInput{
		Name:                         "root_domain",
		AppInputGroupID:              "grp1",
		Source:                       app.AppInputSourceCustomer,
		Type:                         app.AppInputTypeString,
		Default:                      "default.example.com",
		Description:                  "The root domain for the acme-corp install.",
		CloudFormationStackParamName: "InstallRootDomain",
	}
}

func rootDomainInstallInputGroup() app.AppInputGroup {
	return app.AppInputGroup{ID: "grp1", Name: "general"}
}

func TestGetAWSCustomStacksOnlyTemplate_InstallInputNameMatchAutoRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockInstallInputNameMatchTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{Name: "k8s-namespaces", TemplateURL: server.URL + "/stack.yaml", Index: 0},
	})
	inp.AppCfg.InputConfig.AppInputGroups = []app.AppInputGroup{rootDomainInstallInputGroup()}
	inp.AppCfg.InputConfig.AppInputs = []app.AppInput{rootDomainInstallInput()}

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)

	require.Contains(t, tmpl.Parameters, "InstallRootDomain")
	assert.Equal(t, "default.example.com", tmpl.Parameters["InstallRootDomain"].Default)

	stack := tmpl.Resources["K8SNamespaces"]
	nestedStack, ok := stack.(*nestedcloudformation.Stack)
	require.True(t, ok)
	assert.Equal(t, cloudformation.Ref("InstallRootDomain"), nestedStack.Parameters["InstallRootDomain"])
}

func TestGetAWSCustomStacksOnlyTemplate_ExplicitConfigWinsOverInstallInputNameMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockInstallInputNameMatchTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{cfg: &internal.Config{}}
	inp := newCustomStacksOnlyInput([]config.CustomNestedStack{
		{
			Name:        "k8s-namespaces",
			TemplateURL: server.URL + "/stack.yaml",
			Index:       0,
			// Explicit config binds InstallRootDomain to a literal, even though
			// its name also matches a customer-sourced install input's
			// CloudFormationStackParamName -- explicit config must win.
			Parameters: map[string]string{"InstallRootDomain": "acme-corp.example.com"},
		},
	})
	inp.AppCfg.InputConfig.AppInputGroups = []app.AppInputGroup{rootDomainInstallInputGroup()}
	inp.AppCfg.InputConfig.AppInputs = []app.AppInput{rootDomainInstallInput()}
	inp.UnrenderedCustomStackParameters = map[string]map[string]string{
		"k8s-namespaces": {"InstallRootDomain": "acme-corp.example.com"},
	}

	tmpl, err := tpl.getAWSTemplate(inp)
	require.NoError(t, err)

	stack := tmpl.Resources["K8SNamespaces"]
	nestedStack, ok := stack.(*nestedcloudformation.Stack)
	require.True(t, ok)
	assert.Equal(t, "acme-corp.example.com", nestedStack.Parameters["InstallRootDomain"])

	require.Empty(t, ExtractAndStripCustomStacksInputParameters(tmpl))
}
