package config

import (
	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon-go/models"
)

type InstallApprovalOption string

const (
	InstallApprovalOptionApproveAll InstallApprovalOption = "approve-all"
	InstallApprovalOptionPrompt     InstallApprovalOption = "prompt"
	InstallApprovalOptionUnknown    InstallApprovalOption = ""
)

func (o InstallApprovalOption) APIType() models.AppInstallApprovalOption {
	switch o {
	case InstallApprovalOptionApproveAll:
		return models.AppInstallApprovalOptionApproveDashAll
	case InstallApprovalOptionPrompt:
		return models.AppInstallApprovalOptionPrompt
	default:
		// In case for unknown options, default to prompting for approval.
		return models.AppInstallApprovalOptionPrompt
	}
}

type AWSAccount struct {
	Region     string `mapstructure:"region,omitempty"`
	IAMRoleARN string `mapstructure:"iam_role_arn,omitempty"`
}

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name           string                `mapstructure:"name" jsonschema:"required"`
	AWSAccount     *AWSAccount           `mapstructure:"aws_account,omitempty"`
	ApprovalOption InstallApprovalOption `mapstructure:"approval_option,omitempty"`
	Inputs         map[string]string     `mapstructure:"inputs,omitempty"`
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install")
	addDescription(schema, "aws_account", "AWS account related configuration")
	addDescription(schema, "approval_option", "approval option for the install, can be 'approve_all' or 'prompt'")
	addDescription(schema, "inputs", "list of inputs")
}

func (i *Install) parse() error {
	if i == nil {
		return nil
	}

	return nil
}

func (i *Install) Validate() error {
	if i == nil {
		return nil
	}

	return nil
}
