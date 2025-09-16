package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/config/diff"
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

type InputGroup map[string]string

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name           string                `mapstructure:"name" comment:"#:schema https://api.nuon.co/v1/general/config-schema?type=install" jsonschema:"required"`
	ApprovalOption InstallApprovalOption `mapstructure:"approval_option,omitempty"`
	AWSAccount     *AWSAccount           `mapstructure:"aws_account,omitempty"`
	InputGroups    []InputGroup          `mapstructure:"inputs,omitempty"`
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install")
	addDescription(schema, "approval_option", "approval option for the install, can be 'approve_all' or 'prompt'")
	addDescription(schema, "aws_account", "AWS account related configuration")
	addDescription(schema, "inputs", "list of inputs")
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

func (i *Install) ParseIntoInstall(ins *models.AppInstall, inputs *models.AppInstallInputs, cfg *models.AppAppInputConfig, skipSensetive bool) {
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
		if i.InputGroups == nil {
			i.InputGroups = make([]InputGroup, 0)
		}
		if cfg == nil {
			i.InputGroups = append(i.InputGroups, inputs.Values)
		} else {
			groups := make(map[string]map[string]string)

			for _, appInput := range cfg.Inputs {
				if groups[appInput.GroupID] == nil {
					groups[appInput.GroupID] = make(map[string]string)
				}
				if skipSensetive && appInput.Sensitive {
					continue
				}
				groups[appInput.GroupID][appInput.Name] = inputs.Values[appInput.Name]
			}
			for _, group := range groups {
				if len(group) > 0 {
					i.InputGroups = append(i.InputGroups, group)
				}
			}
		}
	}
}

func (i *Install) Diff(upstreamInstall *Install) (string, diff.DiffSummary, error) {
	if i == nil {
		return "", diff.DiffSummary{}, fmt.Errorf("cannot diff a nil install")
	}

	if upstreamInstall == nil {
		upstreamInstall = &Install{
			AWSAccount: &AWSAccount{},
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
			// we skip inputs not present in the upstream state
			// as this only happens for sensetive inputs.
			continue
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
