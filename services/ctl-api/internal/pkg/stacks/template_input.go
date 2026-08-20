package stacks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type TemplateInput struct {
	Install                    *app.Install             `validate:"required"`
	CloudFormationStackVersion *app.InstallStackVersion `validate:"required"`
	InstallState               *state.State             `validate:"required"`
	AppCfg                     *app.AppConfig           `validate:"required"`

	Runner   *app.Runner              `validate:"required"`
	Settings *app.RunnerGroupSettings `validate:"required"`
	APIToken string                   `validate:"required"`

	// ConfiguredRunnerInstanceType is the instance type the app's runner config declares,
	// empty when it declares none. Settings.AWSInstanceType cannot express "unset" — the
	// stack generators resolve the platform default into it before rendering, because the
	// GCP tfvars and the persisted runner group settings both need a concrete value — and
	// the CloudFormation renderer needs the distinction to know whether it may fall back to
	// the nested runner template's own default.
	ConfiguredRunnerInstanceType string

	// runner env vars from runner.toml [env_vars] section, formatted as
	// newline-delimited "export key=value" pairs for injection into user-data.
	RunnerEnvVars string

	// subscripts and embedded templates
	RunnerInitScriptURL string `validate:"required"`

	// AWS-only: the inline source for the phone-home Lambda (Code.ZipFile). Unvalidated
	// because the Azure and GCP paths share this struct — ARM builds its own phone-home
	// script from PhoneHomeURL, and GCP has no equivalent resource. The AWS renderer
	// enforces it instead.
	PhonehomeScript string

	// Custom template URLs for VPC/VNet and runner nested/linked deployments (AWS CloudFormation, Azure ARM)
	VPCNestedStackTemplateURL    string
	RunnerNestedStackTemplateURL string

	// DeploymentScope is the ARM scope the Azure root template renders at, copied
	// from the app's stack config by the caller so the renderer reads one struct.
	// Empty means resource group, which is every install predating the field —
	// compare against app.StackDeploymentScopeSubscription rather than testing for
	// the resource-group value. Unvalidated because the AWS and GCP renderers share
	// this struct and never set it.
	DeploymentScope app.StackDeploymentScope

	// Where the phone-home token map lives, for the Lambda to fetch at invocation
	// time. Never the token itself: the rendered template is fetched
	// unauthenticated from S3 via the quick-link, so it may only carry the secret's
	// location. Empty whenever phone-home auth is not active for the install, which
	// is why these are unvalidated — the GCP path shares this struct and never sets
	// them.
	PhoneHomeSecretARN    string
	PhoneHomeSecretRegion string

	// TargetAWSAccountID pins the rendered CloudFormation template to a single AWS
	// account: the template carries a Rules assertion that fails create/update
	// before any resource is touched when applied in any other account. Set only
	// when the org has phone-home-auth enabled and the install carries a
	// well-formed target account ID — with the flag off the stored value is
	// unvalidated, so pinning to it could brick a stack on junk data. Always empty
	// on the GCP and Azure paths.
	TargetAWSAccountID string
}

// PhoneHomeRoleName is the deterministic IAM role name for an install's phone-home
// Lambda, matching the `<install_id>-<purpose>` convention the install stack's other
// roles already use (`<install_id>-provision`, `-maintenance`, `-deprovision`).
//
// This is the single source of truth for the name. A cross-account grant naming this
// principal cannot be validated at deploy time — IAM accepts a resource policy that
// references a role which does not exist yet, and a mismatch only surfaces as an
// AccessDeniedException at phone-home time. Both the template and any policy that
// names the principal must derive it from here.
func PhoneHomeRoleName(installID string) string {
	return fmt.Sprintf("%s-phone-home", installID)
}

// FormatRunnerEnvVars converts an AppRunnerConfig's EnvVars hstore into a
// newline-delimited string of "export key=value" statements. Default values
// are injected only when not already defined in cfg.EnvVars.
func FormatRunnerEnvVars(cfg *app.AppRunnerConfig, runnerBinaryVersion string) string {
	if cfg == nil {
		cfg = &app.AppRunnerConfig{}
	}

	// Shallow-copy so we don't mutate the caller's map.
	merged := make(map[string]*string, len(cfg.EnvVars)+1)
	for k, v := range cfg.EnvVars {
		merged[k] = v
	}

	// Inject defaults only when not already present.
	if _, ok := merged["RUNNER_BINARY_VERSION"]; !ok {
		merged["RUNNER_BINARY_VERSION"] = &runnerBinaryVersion
	}

	if len(merged) == 0 {
		return ""
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		v := merged[k]
		if v != nil {
			lines = append(lines, fmt.Sprintf("export %s=%s", k, *v))
		}
	}

	return strings.Join(lines, "\n")
}
