package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon-go/models"
	"github.com/nuonco/nuon/pkg/config/diff"
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
	Region string `mapstructure:"region,omitempty" jsonschema:"required"`
}

func (a AWSAccount) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("region").Short("AWS region").
		Long("AWS region where the infrastructure will be deployed").
		Example("us-east-1").
		Example("us-west-2").
		Example("eu-west-1")
}

type InputGroup map[string]string

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name           string                `mapstructure:"name" comment:"#:schema https://api.nuon.co/v1/general/config-schema?type=install" jsonschema:"required"`
	ApprovalOption InstallApprovalOption `mapstructure:"approval_option,omitempty"`
	AWSAccount     *AWSAccount           `mapstructure:"aws_account,omitempty"`
	InputGroups    []InputGroup          `mapstructure:"inputs,omitempty"`
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the install").Required().
		Long("Unique identifier for this install configuration").
		Field("approval_option").Short("approval option for the install").
		Long("Controls how deployments are approved. Options: 'approve-all' (automatic approval) or 'prompt' (requires confirmation)").
		Example("approve-all").
		Example("prompt").
		Field("aws_account").Short("AWS account configuration").
		Long("AWS-specific settings for this install, including region and other account details").
		Field("inputs").Short("input values").
		Long("Array of input groups with key-value pairs for customer inputs provided during installation")
}

func (i *Install) Parse() error {
	if i == nil {
		return nil
	}

	if i.InputGroups == nil {
		i.InputGroups = make([]InputGroup, 0)
	}

	return nil
}

func (i *Install) Validate() error {
	if i == nil {
		return nil
	}

	return nil
}

func (i *Install) FlattenedInputs() map[string]string {
	flattened := make(map[string]string)
	for _, group := range i.InputGroups {
		for key, val := range group {
			flattened[key] = val
		}
	}
	return flattened
}

func (i *Install) Diff(upstreamInstall *Install) (string, diff.DiffSummary, error) {
	if i == nil {
		return "", diff.DiffSummary{}, fmt.Errorf("cannot diff a nil install")
	}

	if upstreamInstall == nil {
		upstreamInstall = &Install{
			AWSAccount:  &AWSAccount{},
			InputGroups: make([]InputGroup, 0),
		}
	}

	diffs := make([]*diff.Diff, 0)
	diffs = append(diffs,
		diff.NewDiff(diff.WithKey("name"), diff.WithStringDiff(upstreamInstall.Name, i.Name)))

	if i.ApprovalOption != InstallApprovalOptionUnknown {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("approval_option"),
			diff.WithStringDiff(string(upstreamInstall.ApprovalOption), string(i.ApprovalOption)),
		))
	}

	if i.AWSAccount != nil {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("aws_account"), diff.WithChildren(diff.NewDiff(
				diff.WithKey("region"),
				diff.WithStringDiff(upstreamInstall.AWSAccount.Region, i.AWSAccount.Region),
			))),
		)
	}

	inputDiffs := make([]*diff.Diff, len(i.InputGroups))
	installInputs := i.FlattenedInputs()
	upstreamInputs := upstreamInstall.FlattenedInputs()

	for key, val := range installInputs {
		current, ok := upstreamInputs[key]
		if !ok {
			// for new installs, upstreamInputs will be empty,
			// this handles the case separately.
			current = ""
		}
		inputDiffs = append(inputDiffs, diff.NewDiff(
			diff.WithKey(key),
			diff.WithStringDiff(current, val),
		))
	}
	diffs = append(diffs, diff.NewDiff(
		diff.WithKey("inputs"),
		diff.WithChildren(inputDiffs...),
	))

	installDiff := diff.NewDiff(
		diff.WithKey(i.Name),
		diff.WithChildren(diffs...),
	)

	return installDiff.String(""), installDiff.Summary(), nil
}
