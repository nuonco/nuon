package config

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/invopop/jsonschema"
)

type CustomNestedStack struct {
	Name         string            `mapstructure:"name" toml:"name" json:"name" jsonschema:"required"`
	TemplateURL  string            `mapstructure:"template_url" toml:"template_url" json:"template_url" jsonschema:"required" features:"template"`
	Index        int               `mapstructure:"index" toml:"index" json:"index" jsonschema:"required"`
	Parameters   map[string]string `mapstructure:"parameters,omitempty" toml:"parameters" json:"parameters,omitempty"`
	Contents     string            `mapstructure:"-" toml:"-" json:"contents,omitempty" jsonschema:"-" features:"get"`
	ContentsHash string            `mapstructure:"-" toml:"-" json:"contents_hash,omitempty" jsonschema:"-"`
}

func (a CustomNestedStack) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("nested stack name").Required().
		Long("Unique name for this custom nested stack. Used as the CloudFormation logical ID and parameter group label.").
		Example("k8s_namespaces").
		Example("eks_access_entries").
		Field("template_url").Short("nested stack template URL").Required().
		Long("URL to the CloudFormation nested template. Parameters are extracted and hoisted into the parent stack.").
		Example("https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/k8s-namespaces.yaml").
		Field("index").Short("execution order index").Required().
		Long("Determines the execution order of custom nested stacks (ascending). Each stack must have a unique index. Lower indices execute first.").
		Example("0").
		Example("1").
		Field("parameters").Short("parameter-to-input mappings").
		Long("Map of CloudFormation parameter names to Nuon install input references. Values must use the template format {{.nuon.install.inputs.<input_name>}}. Only vendor-provided inputs are supported.").
		Example("Namespaces = \"{{.nuon.install.inputs.namespaces}}\"")
}

type StackConfig struct {
	Type        string `mapstructure:"type" toml:"type"`
	Name        string `mapstructure:"name" toml:"name" jsonschema:"required" features:"template"`
	Description string `mapstructure:"description" toml:"description" jsonschema:"required" features:"template"`

	VPCNestedTemplateURL    string `mapstructure:"vpc_nested_template_url" toml:"vpc_nested_template_url" features:"template"`
	RunnerNestedTemplateURL string `mapstructure:"runner_nested_template_url" toml:"runner_nested_template_url" features:"template"`

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
		Long("Name of the CloudFormation stack when deployed in the customer account. Supports Go templating").
		Example("myapp-{{.nuon.install.id}}").
		Example("production-stack").
		Field("description").Short("stack description").Required().
		Long("Description of the stack, displayed in the CloudFormation console. Supports templating").
		Example("Infrastructure stack for MyApp application").
		Field("vpc_nested_template_url").Short("VPC nested template URL").
		Long("URL to the CloudFormation nested template for VPC resources").
		Example("https://s3.amazonaws.com/bucket/vpc-template.yaml").
		Field("runner_nested_template_url").Short("runner nested template URL").
		Long("URL to the CloudFormation nested template for the Nuon runner infrastructure").
		Example("https://s3.amazonaws.com/bucket/runner-template.yaml").
		Field("custom_nested_stacks").Short("custom nested stacks").
		Long("Custom CloudFormation nested stack templates to include. Each entry has a name, template_url, index, and optional parameters. The index field determines execution order (ascending). The parameters field maps CloudFormation parameter names to Nuon install input references using {{.nuon.install.inputs.<name>}} syntax. Remaining parameters are hoisted into a top-level group named after the stack. Executed after first-class nested stacks.").
		Nullable()
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

func ParseInstallInputReference(value string) (string, error) {
	matches := installInputTemplatePattern.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("must be a template reference in the form {{.nuon.install.inputs.<input_name>}}, got %q", value)
	}
	return matches[1], nil
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
	for i, stack := range a.CustomNestedStacks {
		if stack.Name == "" {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks[%d]: name is required", i),
				Err:         fmt.Errorf("custom_nested_stacks[%d]: name is required", i),
			}
		}
		if stack.TemplateURL == "" {
			return ErrConfig{
				Description: fmt.Sprintf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name),
				Err:         fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name),
			}
		}
		for paramName, paramValue := range stack.Parameters {
			if _, err := ParseInstallInputReference(paramValue); err != nil {
				return ErrConfig{
					Description: fmt.Sprintf("custom_nested_stacks[%d] (%s): parameter %q: %s", i, stack.Name, paramName, err),
					Err:         fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q: %w", i, stack.Name, paramName, err),
				}
			}
		}
	}
	return nil
}
