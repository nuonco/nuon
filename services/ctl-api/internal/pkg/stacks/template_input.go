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

	// runner env vars from runner.toml [env_vars] section, formatted as
	// newline-delimited "export key=value" pairs for injection into user-data.
	RunnerEnvVars string

	// subscripts and embedded templates
	RunnerInitScriptURL string `validate:"required"`
	PhonehomeScript     string `validate:"required"`

	// AWS Embeds
	VPCNestedStackTemplateURL    string `validate:"required"`
	RunnerNestedStackTemplateURL string `validate:"required"`
}

// FormatRunnerEnvVars converts an AppRunnerConfig's EnvVars hstore into a
// newline-delimited string of "export key=value" statements.
func FormatRunnerEnvVars(cfg *app.AppRunnerConfig) string {
	if cfg == nil || len(cfg.EnvVars) == 0 {
		return ""
	}

	keys := make([]string, 0, len(cfg.EnvVars))
	for k := range cfg.EnvVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		v := cfg.EnvVars[k]
		if v != nil {
			lines = append(lines, fmt.Sprintf("export %s=%s", k, *v))
		}
	}

	return strings.Join(lines, "\n")
}
