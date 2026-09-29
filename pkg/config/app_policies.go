package config

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"
)

type PoliciesConfig struct {
	Policies []AppPolicy `mapstructure:"policy,omitempty" toml:"policy,omitempty"`

	SourceFile string `mapstructure:"-" toml:"-" json:"-" jsonschema:"-"`
}

func (a *PoliciesConfig) SetSourceFile(path string) {
	a.SourceFile = path
}

func (a *PoliciesConfig) GetSourceFile() string {
	return a.SourceFile
}

func (a PoliciesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("policy").Short("list of policies").
		Long("Array of policy definitions that enforce compliance and security rules across your infrastructure")
}

func (a *PoliciesConfig) parse() error {
	policyLineNumbers := extractPolicyLineNumbers(a.SourceFile)

	for i := range a.Policies {
		if a.Policies[i].Engine == "" {
			a.Policies[i].Engine = AppPolicyEngineKyverno
		}
		if a.Policies[i].SourceFile == "" && a.SourceFile != "" {
			a.Policies[i].SourceFile = a.SourceFile
		}
		if i < len(policyLineNumbers) {
			a.Policies[i].SourceLine = policyLineNumbers[i]
		}
		a.Policies[i].SetNameFromContents()
	}
	return nil
}

func extractPolicyLineNumbers(sourceFile string) []int {
	if sourceFile == "" {
		return nil
	}

	file, err := os.Open(sourceFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lineNumbers []int
	policyHeaderRegex := regexp.MustCompile(`^\s*\[\[\s*policy\s*\]\]\s*$`)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if policyHeaderRegex.MatchString(scanner.Text()) {
			lineNumbers = append(lineNumbers, lineNum)
		}
	}

	return lineNumbers
}

type AppPolicyType string

const (
	// AppPolicyTypeKubernetesCluster applies to kubernetes cluster-level resources (e.g., namespaces, CRDs)
	AppPolicyTypeKubernetesCluster AppPolicyType = "kubernetes_cluster"
	// AppPolicyTypeTerraformModule applies to terraform module components
	AppPolicyTypeTerraformModule AppPolicyType = "terraform_module"
	// AppPolicyTypeHelmChart applies to helm chart components
	AppPolicyTypeHelmChart AppPolicyType = "helm_chart"
	// AppPolicyTypeKubernetesManifest applies to kubernetes manifest components
	AppPolicyTypeKubernetesManifest AppPolicyType = "kubernetes_manifest"
	// AppPolicyTypeDockerBuild applies to docker build components
	AppPolicyTypeDockerBuild AppPolicyType = "docker_build"
	// AppPolicyTypeContainerImage applies to container image components
	AppPolicyTypeContainerImage AppPolicyType = "container_image"
	// AppPolicyTypePulumi applies to pulumi components
	AppPolicyTypePulumi AppPolicyType = "pulumi"
	// AppPolicyTypeSandbox applies to sandbox infrastructure
	AppPolicyTypeSandbox AppPolicyType = "sandbox"
)

var AllAppPolicyTypes = []AppPolicyType{
	AppPolicyTypeKubernetesCluster,
	AppPolicyTypeTerraformModule,
	AppPolicyTypeHelmChart,
	AppPolicyTypeKubernetesManifest,
	AppPolicyTypeDockerBuild,
	AppPolicyTypeContainerImage,
	AppPolicyTypePulumi,
	AppPolicyTypeSandbox,
}

type AppPolicyEngine string

const (
	AppPolicyEngineKyverno AppPolicyEngine = "kyverno"
	AppPolicyEngineOPA     AppPolicyEngine = "opa"
)

var AllAppPolicyEngines = []AppPolicyEngine{
	AppPolicyEngineKyverno,
	AppPolicyEngineOPA,
}

type AppPolicy struct {
	Type       AppPolicyType   `mapstructure:"type"`
	Engine     AppPolicyEngine `mapstructure:"engine,omitempty"`
	Name       string          `mapstructure:"name,omitempty"`
	Contents   string          `mapstructure:"contents" features:"get,template"`
	Components []string        `mapstructure:"components,omitempty"`

	SourceFile string `mapstructure:"-" toml:"-" json:"-" jsonschema:"-"`
	SourceLine int    `mapstructure:"-" toml:"-" json:"-" jsonschema:"-"`
}

func (a *AppPolicy) SetSourceFile(path string) {
	a.SourceFile = path
}

func (a *AppPolicy) GetSourceFile() string {
	return a.SourceFile
}

func (a *AppPolicy) SetSourceLine(line int) {
	a.SourceLine = line
}

func (a *AppPolicy) GetSourceLine() int {
	return a.SourceLine
}

func (a *AppPolicy) SetNameFromSourceFile() {
	if a.Name != "" || a.SourceFile == "" {
		return
	}
	a.Name = extractNameFromPath(a.SourceFile)
}

func (a *AppPolicy) SetNameFromContents() {
	if a.Name != "" || a.Contents == "" {
		return
	}
	if !strings.HasPrefix(a.Contents, "./") && !strings.HasPrefix(a.Contents, "../") && !strings.HasPrefix(a.Contents, "/") {
		return
	}
	a.Name = extractNameFromPath(a.Contents)
}

func extractNameFromPath(path string) string {
	name := path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[:idx]
	}
	return name
}

func (a AppPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("policy type").
		Long("Type of policy that determines where and how it is enforced").
		Enum(
			string(AppPolicyTypeKubernetesCluster),
			string(AppPolicyTypeTerraformModule),
			string(AppPolicyTypeHelmChart),
			string(AppPolicyTypeKubernetesManifest),
			string(AppPolicyTypeContainerImage),
			string(AppPolicyTypePulumi),
			string(AppPolicyTypeSandbox),
		).
		Example("kubernetes_cluster").
		Example("terraform_module").
		Field("engine").Short("policy engine").
		Long("The policy engine used to evaluate the policy. Must be compatible with the policy type.").
		Enum(string(AppPolicyEngineKyverno), string(AppPolicyEngineOPA)).
		Example("kyverno").
		Example("opa").
		Field("name").Short("policy name").
		Long("Human-readable name for the policy. If not specified, will be derived from the source filename when parsing from a policies/ directory.").
		Example("disallow-ingress-nginx-custom-snippets").
		Example("set-karpenter-non-cpu-limits").
		Field("contents").Short("policy document").
		Long("Policy content in the appropriate format for the policy type. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/policy.json), git repositories (git::https://github.com/org/repo//policy.json), file paths (file:///path/to/policy.json), and relative paths (./policy.json)").
		Example("./disallow-ingress-nginx-custom-snippets.yaml").
		Example("./block-mutable-tags.rego").
		Field("components").Short("target components").
		Long("List of component names this policy applies to. Use [\"*\"] to apply to all components of the specified type, or list specific component names. Required for component-scoped policy types - an empty list is rejected. Ignored when type is 'sandbox'.").
		Example("*").
		Example("rds_cluster")
}

type AppPolicyList struct {
	Policy []AppPolicy `mapstructure:"policy" toml:"policy"`
}

func (a AppPolicyList) JSONSchemaExtend(s *jsonschema.Schema) {
	NewSchemaBuilder(s).
		Field("policy").
		Short("list of policy documents").
		Long("One or more AppPolicy objects").
		MinItems(1)
}
