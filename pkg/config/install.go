package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type InstallApprovalOption string

const (
	InstallApprovalOptionApproveAll InstallApprovalOption = "approve-all"
	InstallApprovalOptionAuto       InstallApprovalOption = "auto"
	InstallApprovalOptionPrompt     InstallApprovalOption = "prompt"
	InstallApprovalOptionUnknown    InstallApprovalOption = ""
)

func (o InstallApprovalOption) APIType() models.AppInstallApprovalOption {
	switch o {
	case InstallApprovalOptionApproveAll, InstallApprovalOptionAuto:
		return models.AppInstallApprovalOptionApproveDashAll
	case InstallApprovalOptionPrompt:
		return models.AppInstallApprovalOptionPrompt
	default:
		return models.AppInstallApprovalOptionPrompt
	}
}

type AWSAccount struct {
	Region    string `mapstructure:"region,omitempty" toml:"region,omitempty" jsonschema:"required"`
	AccountID string `mapstructure:"account_id,omitempty" toml:"account_id,omitempty"`
}

func (a AWSAccount) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("region").Short("AWS region").
		Long("AWS region where the infrastructure will be deployed").
		Example("us-east-1").
		Example("us-west-2").
		Example("eu-west-1").
		Field("account_id").Short("AWS account ID").
		Long("AWS account this install targets. Required when phone home authentication is enabled for the organization. Immutable once the install exists.").
		Example("123456789012")
}

type GCPAccount struct {
	ProjectID string `mapstructure:"project_id,omitempty" toml:"project_id,omitempty"`
	Region    string `mapstructure:"region,omitempty" toml:"region,omitempty"`
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

type AzureAccount struct {
	Location       string `mapstructure:"location,omitempty" toml:"location,omitempty" jsonschema:"required"`
	SubscriptionID string `mapstructure:"subscription_id,omitempty" toml:"subscription_id,omitempty"`
}

func (a AzureAccount) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("location").Short("Azure location").
		Long("Azure location/region where the infrastructure will be deployed").
		Example("eastus").
		Example("westus2").
		Example("westeurope").
		Field("subscription_id").Short("Azure subscription ID").
		Long("Azure subscription this install targets. Required when phone home authentication is enabled for the organization. Immutable once the install exists.").
		Example("00000000-0000-0000-0000-000000000000")
}

type InputGroup struct {
	Inputs map[string]string `mapstructure:"inputs" toml:"inputs"`
	Group  string            `mapstructure:"group" toml:"group"`
}

func (ig InputGroup) JSONSchemaExtend(schema *jsonschema.Schema) {
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
	if ig.Inputs == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(ig.Inputs)
}

func (ig *InputGroup) UnmarshalTOML(data []byte) error {
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
		Long("Per-install overrides for custom install stacks. Entries with the same name as app-level stacks replace them; new names are appended. Supports AWS CloudFormation, Azure ARM, and curated GCP modules.").
		Nullable()
}

func (s *InstallStackOverrides) HasOverrides() bool {
	return s != nil && (s.VPCNestedTemplateURL != "" || s.RunnerNestedTemplateURL != "" || len(s.CustomNestedStacks) > 0)
}

type Install struct {
	Name           string                `mapstructure:"name" toml:"name" comment:"install" jsonschema:"required"`
	AppBranch      string                `mapstructure:"app_branch,omitempty" toml:"app_branch,omitempty"`
	AppBranchGroup string                `mapstructure:"app_branch_group,omitempty" toml:"app_branch_group,omitempty"`
	ApprovalOption InstallApprovalOption `mapstructure:"approval_option,omitempty" toml:"approval_option,omitempty"`
	Telemetry      *InstallTelemetry     `mapstructure:"telemetry,omitempty" toml:"telemetry,omitempty" json:",omitempty"`
	Labels         map[string]string     `mapstructure:"labels,omitempty" toml:"labels,omitempty"`
	AWSAccount     *AWSAccount           `mapstructure:"aws_account,omitempty" toml:"aws_account,omitempty"`
	GCPAccount     *GCPAccount           `mapstructure:"gcp_account,omitempty" toml:"gcp_account,omitempty"`
	AzureAccount   *AzureAccount         `mapstructure:"azure_account,omitempty" toml:"azure_account,omitempty"`
	InputGroups    []InputGroup          `mapstructure:"inputs,omitempty" toml:"inputs,omitempty"`

	StackOverrides *InstallStackOverrides `mapstructure:"stack_overrides,omitempty" toml:"stack_overrides,omitempty"`

	ComponentToggles map[string]bool `mapstructure:"component_toggles,omitempty" toml:"component_toggles,omitempty"`

	Components map[string]ComponentOverride `mapstructure:"components,omitempty" toml:"components,omitempty"`
}

type InstallTelemetry struct {
	Enabled *bool `mapstructure:"enabled,omitempty" toml:"enabled,omitempty" json:"enabled,omitempty" extensions:"x-nullable,!x-omitempty"`
}

func (t InstallTelemetry) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("enabled").Short("Enable telemetry").
		Long("Enable or disable telemetry for this install. Omit to preserve the current setting; new installs inherit the organization default.")
}

type ComponentOverride struct {
	HelmValues string `mapstructure:"helm_values,omitempty" toml:"helm_values,omitempty"`
	TFVars     string `mapstructure:"tf_vars,omitempty" toml:"tf_vars,omitempty"`
}

func (c ComponentOverride) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("helm_values").Short("install-level Helm values override (YAML)").
		Long("Raw YAML that deep-merges over the Helm component's app-config values and wins on overlapping keys. Only valid for Helm components.").
		Field("tf_vars").Short("install-level Terraform vars override (.tfvars)").
		Long("Raw .tfvars (HCL or JSON) appended as the final, highest-precedence -var-file. Variables must be declared in the module. Only valid for Terraform components.")
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the install").Required().
		Long("Unique identifier for this install configuration").
		Example("production").
		Example("staging").
		Example("customer-acme").
		Field("app_branch").Short("app branch name or ID").
		Long("App branch this install belongs to, by name or ID. Changing it moves the install and applies the branch's latest run. Required when disable-app-sync is enabled for the organization.").
		Example("main").
		Field("app_branch_group").Short("app branch deployment group").
		Long("Deployment group this install belongs to within its app branch. This explicit selection takes precedence over label matching.").
		Example("canary").
		Field("approval_option").Short("approval option for the install").
		Long("Controls how deployments are approved. Options: 'approve-all' (automatic approval) or 'prompt' (requires confirmation)").
		Example("approve-all").
		Example("prompt").
		Field("telemetry").Short("Install telemetry settings").
		Field("labels").Short("key/value labels for the install").
		Long("Tag installs with arbitrary metadata like environment, region, or version. Values can use the .nuon templating syntax to render from install state, and re-render as state changes.").
		Example(map[string]string{"env": "production", "region": "{{ .nuon.cloud_account.aws.region }}"}).
		Field("aws_account").Short("AWS account configuration").
		Long("AWS-specific settings for this install, including region and other account details").
		Field("gcp_account").Short("GCP account configuration").
		Long("GCP-specific settings for this install, including project ID and region").
		Field("azure_account").Short("Azure account configuration").
		Long("Azure-specific settings for this install, including the deployment location").
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
	if i.AppBranchGroup != "" && i.AppBranch == "" {
		return ErrConfig{
			Description: fmt.Sprintf("install %q: app_branch_group requires app_branch", i.Name),
		}
	}

	// why: Keys are lookup identifiers on every matching surface, so they can never
	// be templated; templated values must parse now, not fail at render time.
	for key, val := range i.Labels {
		if strings.Contains(key, "{{") {
			return ErrConfig{
				Description: fmt.Sprintf("install %q label key %q must not use the interpolation syntax", i.Name, key),
				Err:         fmt.Errorf("install %q label key %q is templated; only label values may be templated", i.Name, key),
			}
		}
		if !labels.IsTemplatedValue(val) {
			continue
		}
		if err := render.ValidateTextTemplate(val); err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("install %q label %q is not a valid template", i.Name, key),
				Err:         fmt.Errorf("install %q label %q template: %w", i.Name, key, err),
			}
		}
	}

	for compName, override := range i.Components {
		if err := ValidateInputValueSyntax(InputTypeYAML, override.HelmValues); err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("component %q helm_values override is not valid YAML", compName),
				Err:         fmt.Errorf("component %q helm_values override: %w", compName, err),
			}
		}
		if err := ValidateInputValueSyntax(InputTypeHCL, override.TFVars); err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("component %q tf_vars override is not valid HCL/tfvars", compName),
				Err:         fmt.Errorf("component %q tf_vars override: %w", compName, err),
			}
		}
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
	for compName, override := range i.Components {
		if override.HelmValues != "" {
			flattened[HelmValuesOverrideInputName(compName)] = override.HelmValues
		}
		if override.TFVars != "" {
			flattened[TFVarsOverrideInputName(compName)] = override.TFVars
		}
	}
	for compName, enabled := range i.ComponentToggles {
		flattened[EnabledOverrideInputName(compName)] = strconv.FormatBool(enabled)
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

	if i.AppBranch != "" {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("app_branch"),
			diff.WithStringDiff(upstreamInstall.AppBranch, i.AppBranch),
		))
	}
	if i.AppBranchGroup != "" || upstreamInstall.AppBranchGroup != "" {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("app_branch_group"),
			diff.WithStringDiff(upstreamInstall.AppBranchGroup, i.AppBranchGroup),
		))
	}

	if i.ApprovalOption != InstallApprovalOptionUnknown {
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("approval_option"),
			diff.WithStringDiff(string(upstreamInstall.ApprovalOption), string(i.ApprovalOption)),
		))
	}

	if i.Telemetry != nil && i.Telemetry.Enabled != nil {
		previous := "inherit"
		if upstreamInstall.Telemetry != nil && upstreamInstall.Telemetry.Enabled != nil {
			previous = strconv.FormatBool(*upstreamInstall.Telemetry.Enabled)
		}
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("telemetry.enabled"),
			diff.WithStringDiff(previous, strconv.FormatBool(*i.Telemetry.Enabled)),
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

	if len(i.Labels) > 0 || len(upstreamInstall.Labels) > 0 {
		labelDiffs := make([]*diff.Diff, 0)
		seen := make(map[string]bool)
		for k, v := range i.Labels {
			seen[k] = true
			labelDiffs = append(labelDiffs, diff.NewDiff(
				diff.WithKey(k),
				diff.WithStringDiff(upstreamInstall.Labels[k], v),
			))
		}
		for k, v := range upstreamInstall.Labels {
			if seen[k] {
				continue
			}
			labelDiffs = append(labelDiffs, diff.NewDiff(
				diff.WithKey(k),
				diff.WithStringDiff(v, ""),
			))
		}
		if len(labelDiffs) > 0 {
			diffs = append(diffs, diff.NewDiff(
				diff.WithKey("labels"),
				diff.WithChildren(labelDiffs...),
			))
		}
	}

	if i.AWSAccount != nil {
		upstreamAWS := &AWSAccount{}
		if upstreamInstall.AWSAccount != nil {
			upstreamAWS = upstreamInstall.AWSAccount
		}
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("aws_account"), diff.WithChildren(
				diff.NewDiff(diff.WithKey("region"), diff.WithStringDiff(upstreamAWS.Region, i.AWSAccount.Region)),
				diff.NewDiff(diff.WithKey("account_id"), diff.WithStringDiff(upstreamAWS.AccountID, i.AWSAccount.AccountID)),
			)),
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

	if i.AzureAccount != nil {
		upstreamAzure := &AzureAccount{}
		if upstreamInstall.AzureAccount != nil {
			upstreamAzure = upstreamInstall.AzureAccount
		}
		diffs = append(diffs, diff.NewDiff(
			diff.WithKey("azure_account"), diff.WithChildren(
				diff.NewDiff(diff.WithKey("location"), diff.WithStringDiff(upstreamAzure.Location, i.AzureAccount.Location)),
				diff.NewDiff(diff.WithKey("subscription_id"), diff.WithStringDiff(upstreamAzure.SubscriptionID, i.AzureAccount.SubscriptionID)),
			)),
		)
	}

	inputDiffs := make([]*diff.Diff, len(i.InputGroups))
	installInputs := i.FlattenedInputs()
	upstreamInputs := upstreamInstall.FlattenedInputs()

	for key, val := range installInputs {
		inputDiffs = append(inputDiffs, inputDiffNode(key, upstreamInputs[key], val))
	}

	// why: Detect cleared component overrides: present upstream but no longer set
	// locally. Normal inputs are additive (omitting one does not clear it), so
	// this removal pass is scoped to reserved synthetic override keys, which
	// revert their component to app-config values when set back to empty.
	for key, current := range upstreamInputs {
		if _, ok := installInputs[key]; ok {
			continue
		}
		if !IsComponentOverrideInputName(key) {
			continue
		}
		inputDiffs = append(inputDiffs, inputDiffNode(key, current, ""))
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
