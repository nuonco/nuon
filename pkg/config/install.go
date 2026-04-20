package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pelletier/go-toml/v2"
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
	Region string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
}

func (a AWSAccount) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("region").Short("AWS region").
		Long("AWS region where the infrastructure will be deployed").
		Example("us-east-1").
		Example("us-west-2").
		Example("eu-west-1")
}

type GCPAccount struct {
	ProjectID string `mapstructure:"project_id,omitempty" toml:"project_id,omitempty" jsonschema:"required"`
	Region    string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
}

func (a GCPAccount) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("project_id").Short("GCP project ID").
		Long("GCP project where the infrastructure will be deployed").
		Example("my-gcp-project").
		Field("region").Short("GCP region").
		Long("GCP region where the infrastructure will be deployed").
		Example("us-central1").
		Example("europe-west1")
}

type InputGroup struct {
	// mapstructure is able to decode map into Inputgroup because the type of InputGroup.Inputs matches that of what
	// expected by mapstructure.
	Inputs map[string]string `mapstructure:"inputs" toml:"inputs"`
	// this property should only be used for writing comment for input group, and should not be used anywhere else.
	Group string `mapstructure:"group" toml:"group"`
}

func (ig InputGroup) JSONSchemaExtend(schema *jsonschema.Schema) {
	// Make schema treat InputGroup as a map (additionalProperties pattern)
	schema.Type = "object"
	schema.AdditionalProperties = &jsonschema.Schema{
		Type: "string",
	}
	schema.Properties = nil
	schema.Required = nil
}

func (ig InputGroup) MarshalTOML() ([]byte, error) {
	return toml.Marshal(ig.Inputs)
}

func (ig InputGroup) MarshalJSON() ([]byte, error) {
	// Marshal as flat map to match the JSONSchemaExtend definition
	if ig.Inputs == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(ig.Inputs)
}

func (ig *InputGroup) UnmarshalTOML(data []byte) error {
	// First unmarshal the TOML data into a map
	var m map[string]string
	if err := toml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("failed to unmarshal TOML data: %w", err)
	}

	ig.Inputs = map[string]string{}
	for k, v := range m {
		ig.Inputs[k] = fmt.Sprint(v)
	}
	return nil
}

func (ig InputGroup) TOMLComment() string {
	return fmt.Sprintf("input.group: %s", ig.Group)
}

// InstallStackOverrides holds per-install overrides for the app-level stack
// template configuration. Nil fields mean "use the app default".
type InstallStackOverrides struct {
	VPCNestedTemplateURL    string              `mapstructure:"vpc_nested_template_url,omitempty" toml:"vpc_nested_template_url,omitempty"`
	RunnerNestedTemplateURL string              `mapstructure:"runner_nested_template_url,omitempty" toml:"runner_nested_template_url,omitempty"`
	CustomNestedStacks      []CustomNestedStack `mapstructure:"custom_nested_stacks,omitempty" toml:"custom_nested_stacks,omitempty"`
}

func (a InstallStackOverrides) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("vpc_nested_template_url").Short("VPC nested template URL override").
		Long("Per-install override for the VPC nested CloudFormation template URL. Overrides the app-level default from stack.toml.").
		Example("https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/custom-vpc.yaml").
		Field("runner_nested_template_url").Short("Runner nested template URL override").
		Long("Per-install override for the runner nested CloudFormation template URL. Overrides the app-level default from stack.toml.").
		Example("https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/custom-runner.yaml").
		Field("custom_nested_stacks").Short("Custom nested stack overrides").
		Long("Per-install overrides for custom nested CloudFormation stacks. Entries with the same name as app-level stacks replace them; new names are appended.").
		Nullable()
}

// HasOverrides returns true when any override field is set.
func (s *InstallStackOverrides) HasOverrides() bool {
	return s != nil && (s.VPCNestedTemplateURL != "" || s.RunnerNestedTemplateURL != "" || len(s.CustomNestedStacks) > 0)
}

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name           string                `mapstructure:"name" toml:"name" comment:"install" jsonschema:"required"`
	ApprovalOption InstallApprovalOption `mapstructure:"approval_option,omitempty" toml:"approval_option,omitempty"`
	AWSAccount     *AWSAccount           `mapstructure:"aws_account,omitempty" toml:"aws_account,omitempty"`
	GCPAccount     *GCPAccount           `mapstructure:"gcp_account,omitempty" toml:"gcp_account,omitempty"`
	InputGroups    []InputGroup          `mapstructure:"inputs,omitempty" toml:"inputs,omitempty"`

	StackOverrides *InstallStackOverrides `mapstructure:"stack_overrides,omitempty" toml:"stack_overrides,omitempty"`
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the install").Required().
		Long("Unique identifier for this install configuration").
		Example("production").
		Example("staging").
		Example("customer-acme").
		Field("approval_option").Short("approval option for the install").
		Long("Controls how deployments are approved. Options: 'approve-all' (automatic approval) or 'prompt' (requires confirmation)").
		Example("approve-all").
		Example("prompt").
		Field("aws_account").Short("AWS account configuration").
		Long("AWS-specific settings for this install, including region and other account details").
		Field("gcp_account").Short("GCP account configuration").
		Long("GCP-specific settings for this install, including project ID and region").
		Field("inputs").Short("input values").
		Long("Array of input groups with key-value pairs for customer inputs provided during installation").
		Type("array").
		Field("stack_overrides").Short("Stack template overrides").
		Long("Per-install overrides for the app-level stack template configuration. Overrides take precedence over app-level defaults.")
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
		for key, val := range group.Inputs {
			flattened[key] = val
		}
	}
	return flattened
}

func (i *Install) Diff(upstreamInstall *Install) (*diff.Diff, error) {
	if i == nil {
		return nil, fmt.Errorf("cannot diff a nil install")
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

	if i.StackOverrides.HasOverrides() || upstreamInstall.StackOverrides.HasOverrides() {
		curr := i.StackOverrides
		if curr == nil {
			curr = &InstallStackOverrides{}
		}
		upstream := upstreamInstall.StackOverrides
		if upstream == nil {
			upstream = &InstallStackOverrides{}
		}

		var stackDiffs []*diff.Diff
		if curr.VPCNestedTemplateURL != "" || upstream.VPCNestedTemplateURL != "" {
			stackDiffs = append(stackDiffs, diff.NewDiff(
				diff.WithKey("vpc_nested_template_url"),
				diff.WithStringDiff(upstream.VPCNestedTemplateURL, curr.VPCNestedTemplateURL),
			))
		}
		if curr.RunnerNestedTemplateURL != "" || upstream.RunnerNestedTemplateURL != "" {
			stackDiffs = append(stackDiffs, diff.NewDiff(
				diff.WithKey("runner_nested_template_url"),
				diff.WithStringDiff(upstream.RunnerNestedTemplateURL, curr.RunnerNestedTemplateURL),
			))
		}

		// Diff custom nested stacks by name.
		upstreamByName := make(map[string]CustomNestedStack, len(upstream.CustomNestedStacks))
		for _, s := range upstream.CustomNestedStacks {
			upstreamByName[s.Name] = s
		}
		seen := make(map[string]bool)
		for _, s := range curr.CustomNestedStacks {
			seen[s.Name] = true
			if us, ok := upstreamByName[s.Name]; ok {
				if s.TemplateURL != us.TemplateURL {
					stackDiffs = append(stackDiffs, diff.NewDiff(
						diff.WithKey(s.Name+".template_url"),
						diff.WithStringDiff(us.TemplateURL, s.TemplateURL),
					))
				}
			} else {
				stackDiffs = append(stackDiffs, diff.NewDiff(
					diff.WithKey(s.Name),
					diff.WithStringDiff("", s.TemplateURL),
				))
			}
		}
		for _, s := range upstream.CustomNestedStacks {
			if !seen[s.Name] {
				stackDiffs = append(stackDiffs, diff.NewDiff(
					diff.WithKey(s.Name),
					diff.WithStringDiff(s.TemplateURL, ""),
				))
			}
		}

		if len(stackDiffs) > 0 {
			diffs = append(diffs, diff.NewDiff(
				diff.WithKey("stack_overrides"),
				diff.WithChildren(stackDiffs...),
			))
		}
	}

	if i.AWSAccount != nil {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("aws_account"), diff.WithChildren(diff.NewDiff(
				diff.WithKey("region"),
				diff.WithStringDiff(upstreamInstall.AWSAccount.Region, i.AWSAccount.Region),
			))),
		)
	}

	if i.GCPAccount != nil {
		upstreamGCP := &GCPAccount{}
		if upstreamInstall.GCPAccount != nil {
			upstreamGCP = upstreamInstall.GCPAccount
		}
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("gcp_account"), diff.WithChildren(
				diff.NewDiff(diff.WithKey("project_id"), diff.WithStringDiff(upstreamGCP.ProjectID, i.GCPAccount.ProjectID)),
				diff.NewDiff(diff.WithKey("region"), diff.WithStringDiff(upstreamGCP.Region, i.GCPAccount.Region)),
			)),
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

	return installDiff, nil
}
