package config

import (
	"github.com/invopop/jsonschema"
)

type StackConfig struct {
	Type        string `mapstructure:"type"`
	Name        string `mapstructure:"name" jsonschema:"required" features:"template"`
	Description string `mapstructure:"description" jsonschema:"required" features:"template"`

	VPCNestedTemplateURL    string `mapstructure:"vpc_nested_template_url"`
	RunnerNestedTemplateURL string `mapstructure:"runner_nested_template_url"`
}

func (a StackConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("stack type").
		Long("Type of infrastructure stack. Currently only 'aws-cloudformation' is supported").
		Example("aws-cloudformation").
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
		Example("https://s3.amazonaws.com/bucket/runner-template.yaml")
}

func (a *StackConfig) parse() error {
	return nil
}
