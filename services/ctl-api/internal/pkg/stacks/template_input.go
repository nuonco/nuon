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
