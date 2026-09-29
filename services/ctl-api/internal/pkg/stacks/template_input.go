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

	ConfiguredRunnerInstanceType string

	RunnerEnvVars string

	RunnerInitScriptURL string `validate:"required"`

	PhonehomeScript string

	VPCNestedStackTemplateURL    string
	RunnerNestedStackTemplateURL string

	DeploymentScope app.StackDeploymentScope

	PhoneHomeSecretARN    string
	PhoneHomeSecretRegion string

	PhoneHomeIdentityName string

	CustomStacksOnly bool

	UnrenderedCustomStackParameters map[string]map[string]string
}

// why: PhoneHomeRoleName is the deterministic IAM role name for an install's phone-home
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

func FormatRunnerEnvVars(cfg *app.AppRunnerConfig, runnerBinaryVersion string) string {
	if cfg == nil {
		cfg = &app.AppRunnerConfig{}
	}

	merged := make(map[string]*string, len(cfg.EnvVars)+1)
	for k, v := range cfg.EnvVars {
		merged[k] = v
	}

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
