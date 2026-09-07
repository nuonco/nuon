package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
)

// CustomNestedStackStatus describes whether a custom nested stack's template
// contents have been uploaded to the managed S3 bucket and are ready to be
// referenced in a generated CloudFormation/ARM stack.
type CustomNestedStackStatus string

const (
	// CustomNestedStackStatusPending indicates the template contents have been
	// received but not yet uploaded to S3 (ContentsHash not yet set).
	CustomNestedStackStatusPending CustomNestedStackStatus = "pending"
	// CustomNestedStackStatusReady indicates the template contents have been
	// uploaded to S3 and ContentsHash is set; the stack can be generated.
	CustomNestedStackStatusReady CustomNestedStackStatus = "ready"
	// CustomNestedStackStatusError indicates uploading the template contents failed.
	CustomNestedStackStatusError CustomNestedStackStatus = "error"
)

type CustomNestedStack struct {
	Name         string                  `mapstructure:"name" toml:"name" json:"name" jsonschema:"required"`
	TemplateURL  string                  `mapstructure:"template_url" toml:"template_url" json:"template_url" jsonschema:"required" features:"template"`
	Index        int                     `mapstructure:"index" toml:"index" json:"index" jsonschema:"required"`
	Parameters   map[string]string       `mapstructure:"parameters,omitempty" toml:"parameters" json:"parameters,omitempty"`
	Contents     string                  `mapstructure:"-" toml:"-" json:"contents,omitempty" jsonschema:"-" features:"get"`
	ContentsHash string                  `mapstructure:"-" toml:"-" json:"contents_hash,omitempty" jsonschema:"-"`
	Status       CustomNestedStackStatus `mapstructure:"-" toml:"-" json:"status,omitempty" jsonschema:"-"`
	// TemplateSourceURL is the public URL of the uploaded template contents.
	// TemplateURL is whatever the vendor wrote in their config — usually a path
	// relative to the config dir — so it is not resolvable by anything that did
	// not do the original parse. Set when the contents are uploaded; empty for
	// configs synced before this field existed.
	TemplateSourceURL string `mapstructure:"-" toml:"-" json:"template_source_url,omitempty" jsonschema:"-"`
}

func (a CustomNestedStack) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("nested stack name").Required().
		Long("Stable, unique name for the custom stack. Used for deployment naming and output lookup.").
		Example("k8s_namespaces").
		Example("eks_access_entries").
		Field("template_url").Short("custom stack template or module").Required().
		Long("AWS CloudFormation template URL or relative path, Azure compiled ARM JSON URL or relative path, or GCP curated module path.").
		Example("https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/k8s-namespaces.yaml").
		Example("./arm/storage.json").
		Example("github.com/nuonco/install-stacks//gcp/modules/bucket").
		Field("index").Short("execution order index").Required().
		Long("Unique ordering key. AWS and Azure custom stacks execute in ascending order.").
		Example("0").
		Example("1").
		Field("parameters").Short("parameter values").
		Long("Map of nested stack parameter names to templated values. Values are rendered when the install stack is generated, so they may only reference state that exists at that point: " + availableStackParameterRefs + ". Sandbox, component, action and install stack outputs are not available.").
		Example("Namespaces = \"{{.nuon.install.inputs.namespaces}}\"").
		Example("RootDomain = \"{{ if .nuon.install.inputs.root_domain }}{{ .nuon.install.inputs.root_domain }}{{ else }}{{ .nuon.install.id }}.example.com{{ end }}\"")
}

// GCPCustomStackModuleMarker separates the repo address from the curated
// module name in a gcp-terraform custom stack template_url. Any repo works —
// forks carry their own modules — as long as the install stack the customer
// applies comes from the same repo; the rendered tfvars only carry the name.
const GCPCustomStackModuleMarker = "//gcp/modules/"

// GCPCustomStackModulePrefix is the canonical upstream form, used in examples
// and error messages.
const GCPCustomStackModulePrefix = "github.com/nuonco/install-stacks" + GCPCustomStackModuleMarker

// GCPModuleName returns the curated module name a gcp-terraform custom stack
// references, or "" when the template_url is not a gcp modules path.
func (c CustomNestedStack) GCPModuleName() string {
	idx := strings.LastIndex(c.TemplateURL, GCPCustomStackModuleMarker)
	if idx <= 0 {
		return ""
	}
	name := c.TemplateURL[idx+len(GCPCustomStackModuleMarker):]
	if name == "" || strings.ContainsAny(name, "/?") {
		return ""
	}
	return name
}

// ValidateGCPCustomNestedStacks validates requirements that are known before
// Terraform runs in the customer's project.
func ValidateGCPCustomNestedStacks(stackType string, stacks []CustomNestedStack) error {
	if stackType != "gcp-terraform" {
		return nil
	}

	for i, stack := range stacks {
		moduleName := stack.GCPModuleName()
		if moduleName == "" {
			msg := fmt.Sprintf("custom_nested_stacks[%d] (%s): gcp-terraform custom stacks must reference a gcp modules path (<repo>%s<name>), e.g. %s<name>", i, stack.Name, GCPCustomStackModuleMarker, GCPCustomStackModulePrefix)
			return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
		}
		if moduleName == "dns" && strings.TrimSpace(stack.Parameters["dns_name"]) == "" {
			msg := fmt.Sprintf("custom_nested_stacks[%d] (%s): parameters.dns_name is required for the GCP dns module", i, stack.Name)
			return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
		}
	}

	return nil
}

// Deployment scopes for the generated Azure install stack root template.
//
// The empty string is equivalent to StackDeploymentScopeResourceGroup and is
// deliberately never normalized on write: configs stored before this field
// existed would otherwise show a spurious diff on the next sync.
const (
	// StackDeploymentScopeResourceGroup confines every resource to the install's
	// own resource group. This is the default and the only behaviour before the
	// field existed.
	StackDeploymentScopeResourceGroup = "resource_group"
	// StackDeploymentScopeSubscription deploys the root template at subscription
	// scope, which is what lets nested stacks create their own resource groups.
	StackDeploymentScopeSubscription = "subscription"
)

type StackConfig struct {
	Type        string `mapstructure:"type" toml:"type"`
	Name        string `mapstructure:"name" toml:"name" jsonschema:"required" features:"template"`
	Description string `mapstructure:"description" toml:"description" jsonschema:"required" features:"template"`

	VPCNestedTemplateURL    string `mapstructure:"vpc_nested_template_url" toml:"vpc_nested_template_url" features features:"template"`
	RunnerNestedTemplateURL string `mapstructure:"runner_nested_template_url" toml:"runner_nested_template_url" features features:"template"`

	DeploymentScope string `mapstructure:"deployment_scope" toml:"deployment_scope,omitempty"`

	CustomNestedStacks []CustomNestedStack `mapstructure:"custom_nested_stacks" toml:"custom_nested_stacks"`
}

func (a StackConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("stack type").
		Long("Type of infrastructure stack. Supported values: 'aws-cloudformation', 'azure-bicep' (Azure), 'gcp-terraform' (Google Cloud).").
		Example("aws-cloudformation").
		Example("azure-bicep").
		Example("gcp-terraform").
		Field("name").Short("stack name").Required().
		Long("Name of the install stack when deployed in the customer account or project. Supports Go templating").
		Example("myapp-{{.nuon.install.id}}").
		Example("production-stack").
		Field("description").Short("stack description").Required().
		Long("Description of the install stack. Supports Go templating").
		Example("Infrastructure stack for MyApp application").
		Field("vpc_nested_template_url").Short("VPC nested template URL").
		Long("URL to the CloudFormation nested template for VPC resources").
		Example("https://s3.amazonaws.com/bucket/vpc-template.yaml").
		Field("runner_nested_template_url").Short("runner nested template URL").
		Long("URL to the CloudFormation nested template for the Nuon runner infrastructure").
		Example("https://s3.amazonaws.com/bucket/runner-template.yaml").
		Field("deployment_scope").Short("deployment scope").
		Long("Scope the generated install stack root template deploys at. Only supported for 'azure-bicep'. Supported values: 'resource_group' (the default) confines every resource to the install's own resource group; 'subscription' deploys at subscription scope, which lets nested stack templates create their own resource groups — needed when an app splits networking, application and security resources across several groups.").
		Example("subscription").
		Field("custom_nested_stacks").Short("custom nested stacks").
		Long("Custom install-stack resources to include. Each entry has a name, template_url, index, and optional parameters. AWS uses CloudFormation templates, Azure uses compiled ARM JSON, and GCP uses curated install-stack modules.").
		Nullable()
}

// ValidateDeploymentScope checks a stack's deployment_scope against its type.
// Shared by StackConfig.parse (the TOML path) and build.StackConfig (the API and
// syncer paths) so every entry point rejects the same input.
func ValidateDeploymentScope(scope, stackType string) error {
	switch scope {
	case "", StackDeploymentScopeResourceGroup:
		return nil
	case StackDeploymentScopeSubscription:
		if stackType != "azure-bicep" {
			msg := fmt.Sprintf("deployment_scope %q is only supported when type is azure-bicep, got %q", scope, stackType)
			return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
		}
		return nil
	default:
		msg := fmt.Sprintf("deployment_scope must be %q or %q, got %q", StackDeploymentScopeResourceGroup, StackDeploymentScopeSubscription, scope)
		return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
	}
}

func ValidateTemplateURL(templateURL string, fieldName string) error {
	u, err := url.Parse(templateURL)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("%s is not a valid URL: %s", fieldName, err),
			Err:         fmt.Errorf("%s: %w", fieldName, err),
		}
	}
	if u.Scheme == "" || u.Host == "" {
		return ErrConfig{
			Description: fmt.Sprintf("%s must be a valid URL with scheme and host (e.g. https://s3.amazonaws.com/bucket/template.yaml), got %q", fieldName, templateURL),
			Err:         fmt.Errorf("%s: missing scheme or host", fieldName),
		}
	}
	if !isS3URL(u) {
		return ErrConfig{
			Description: fmt.Sprintf("%s must be an S3 URL (e.g. https://s3.amazonaws.com/bucket/template.yaml or https://bucket.s3.region.amazonaws.com/key), got %q", fieldName, templateURL),
			Err:         fmt.Errorf("%s: not an S3 URL", fieldName),
		}
	}
	return nil
}

func ValidateHTTPSURL(templateURL string, fieldName string) error {
	u, err := url.Parse(templateURL)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("%s is not a valid URL: %s", fieldName, err),
			Err:         fmt.Errorf("%s: %w", fieldName, err),
		}
	}
	if u.Scheme != "https" || u.Host == "" {
		return ErrConfig{
			Description: fmt.Sprintf("%s must be a valid HTTPS URL (e.g. https://example.com/template.json), got %q", fieldName, templateURL),
			Err:         fmt.Errorf("%s: must be an HTTPS URL", fieldName),
		}
	}
	return nil
}

var installInputTemplatePattern = regexp.MustCompile(`^\{\{\s*\.nuon\.install\.inputs\.([a-zA-Z0-9_]+)\s*\}\}$`)

// ParseInstallInputReference matches the single-install-input reference form,
// {{.nuon.install.inputs.<input_name>}}, and returns the input name.
//
// Custom nested stack parameters are no longer limited to this form -- they are
// rendered as full templates before the stack renderers see them (see
// ValidateStackParameterTemplate). The stack renderers keep this as a fallback for
// call paths that read config without rendering it first.
func ParseInstallInputReference(value string) (string, error) {
	matches := installInputTemplatePattern.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("must be a template reference in the form {{.nuon.install.inputs.<input_name>}}, got %q", value)
	}
	return matches[1], nil
}

// RenderCustomNestedStackParameters renders each stack's parameter values in place.
//
// Parameters are rendered here rather than through the features:"template" tag on
// the field because RenderStruct routes tagged fields through html/template, which
// would escape "&" and quotes in a value that is bound for a cloud provider's API.
func RenderCustomNestedStackParameters(stacks []CustomNestedStack, data map[string]any) error {
	for i := range stacks {
		if err := render.RenderTextStringMap(stacks[i].Parameters, data); err != nil {
			return fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stacks[i].Name, err)
		}
	}

	return nil
}

// Custom nested stack parameters are rendered when the install stack is generated,
// which happens before the customer applies the stack. References to anything that
// only exists after that point cannot resolve, and because the renderer runs with
// missingkey=error they do not render empty -- they fail the generate-install-stack
// workflow at install create. Reject them at sync instead, where the vendor sees it.
var disallowedStackParameterRefTypes = []refs.RefType{
	refs.RefTypeSandbox,
	refs.RefTypeComponents,
	refs.RefTypeActions,
	refs.RefTypeInstallStack,
}

// Legacy spellings of the same late-bound state. refs.ParseFieldRefs' patterns only
// cover the flattened paths, so these are matched literally.
var disallowedStackParameterPaths = []string{
	"nuon.install.sandbox.",
	"nuon.install.components.",
	"nuon.install.actions.",
}

const availableStackParameterRefs = ".nuon.install.inputs.*, .nuon.inputs.inputs.*, .nuon.install.id, .nuon.app.*, .nuon.org.*"

// ValidateStackParameterTemplate validates a custom nested stack parameter value: it
// must be a parseable template that only references state which is populated when the
// install stack is generated.
func ValidateStackParameterTemplate(value string) error {
	if value == "" {
		return fmt.Errorf("must not be empty")
	}

	if err := render.ValidateTextTemplate(value); err != nil {
		return fmt.Errorf("is not a valid template: %w", err)
	}

	for _, ref := range refs.ParseFieldRefs(value) {
		if generics.SliceContains(ref.Type, disallowedStackParameterRefTypes) {
			return fmt.Errorf("references %s, which is not populated when the install stack is generated; available: %s", ref.Input, availableStackParameterRefs)
		}
	}

	for _, path := range disallowedStackParameterPaths {
		if strings.Contains(value, path) {
			return fmt.Errorf("references %s*, which is not populated when the install stack is generated; available: %s", path, availableStackParameterRefs)
		}
	}

	return nil
}

var s3HostPattern = regexp.MustCompile(
	`^(.+\.)?s3([.-][a-z0-9-]+)?\.amazonaws\.com$`,
)

func isS3URL(u *url.URL) bool {
	if u.Scheme != "https" {
		return false
	}
	return s3HostPattern.MatchString(u.Host)
}

func (a *StackConfig) parse() error {
	if err := ValidateDeploymentScope(a.DeploymentScope, a.Type); err != nil {
		return err
	}
	if a.Type == "aws-cloudformation" {
		if a.VPCNestedTemplateURL == "" {
			return ErrConfig{
				Description: "vpc_nested_template_url is required when type is aws-cloudformation",
				Err:         fmt.Errorf("vpc_nested_template_url is required when type is aws-cloudformation"),
			}
		}
		if a.RunnerNestedTemplateURL == "" {
			return ErrConfig{
				Description: "runner_nested_template_url is required when type is aws-cloudformation",
				Err:         fmt.Errorf("runner_nested_template_url is required when type is aws-cloudformation"),
			}
		}
	}
	if a.VPCNestedTemplateURL != "" {
		if a.Type == "azure-bicep" {
			if err := ValidateHTTPSURL(a.VPCNestedTemplateURL, "vpc_nested_template_url"); err != nil {
				return err
			}
		} else {
			if err := ValidateTemplateURL(a.VPCNestedTemplateURL, "vpc_nested_template_url"); err != nil {
				return err
			}
		}
	}
	if a.RunnerNestedTemplateURL != "" {
		if a.Type == "azure-bicep" {
			if err := ValidateHTTPSURL(a.RunnerNestedTemplateURL, "runner_nested_template_url"); err != nil {
				return err
			}
		} else {
			if err := ValidateTemplateURL(a.RunnerNestedTemplateURL, "runner_nested_template_url"); err != nil {
				return err
			}
		}
	}
	if err := ValidateGCPCustomNestedStacks(a.Type, a.CustomNestedStacks); err != nil {
		return err
	}
	seenIndices := map[int]string{}
	seenNames := map[string]int{}
	for i, stack := range a.CustomNestedStacks {
		if stack.Name == "" {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks[%d]: name is required", i),
				Err:         fmt.Errorf("custom_nested_stacks[%d]: name is required", i),
			}
		}
		if prev, exists := seenNames[stack.Name]; exists {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks[%d] (%s): name is already used by custom_nested_stacks[%d]; each stack must have a unique name", i, stack.Name, prev),
				Err:         fmt.Errorf("custom_nested_stacks[%d] (%s): name is already used by custom_nested_stacks[%d]; each stack must have a unique name", i, stack.Name, prev),
			}
		}
		seenNames[stack.Name] = i
		if stack.TemplateURL == "" {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name),
				Err:         fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name),
			}
		}
		if prev, exists := seenIndices[stack.Index]; exists {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks: index %d is used by both %q and %q; each stack must have a unique index", stack.Index, prev, stack.Name),
				Err:         fmt.Errorf("custom_nested_stacks: index %d is used by both %q and %q; each stack must have a unique index", stack.Index, prev, stack.Name),
			}
		}
		seenIndices[stack.Index] = stack.Name
		for paramName, paramValue := range stack.Parameters {
			if err := ValidateStackParameterTemplate(paramValue); err != nil {
				return ErrConfig{
					Description: fmt.Sprintf("custom_nested_stacks[%d] (%s): parameter %q: %s", i, stack.Name, paramName, err),
					Err:         fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q: %w", i, stack.Name, paramName, err),
				}
			}
		}
	}
	return nil
}
