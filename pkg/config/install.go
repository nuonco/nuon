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
	Region string `mapstructure:"region,omitempty"`
}

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name           string                `mapstructure:"name" comment:"#:schema https://api.nuon.co/v1/general/config-schema?type=install" jsonschema:"required"`
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

func (i *Install) Parse() error {
	if i == nil {
		return nil
	}

	if i.Inputs == nil {
		i.Inputs = make(map[string]string)
	}

	return nil
}

func (i *Install) Validate() error {
	if i == nil {
		return nil
	}

	return nil
}

func (i *Install) ParseInstall(ins *models.AppInstall, inputs *models.AppInstallInputs) {
	if ins != nil {
		i.Name = ins.Name
		if ins.AwsAccount != nil {
			i.AWSAccount = &AWSAccount{
				Region: ins.AwsAccount.Region,
			}
		}
		if ins.InstallConfig != nil {
			i.ApprovalOption = InstallApprovalOption(ins.InstallConfig.ApprovalOption)
		}
	}
	if inputs != nil {
		i.Inputs = inputs.Values
	}
}
