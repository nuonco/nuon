package validate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/urlpolicy"
)

func testURLPolicy(t *testing.T, mode urlpolicy.Mode) *urlpolicy.Policy {
	t.Helper()
	policy, err := urlpolicy.New(urlpolicy.Config{
		Mode: mode,
		Categories: map[urlpolicy.Category]urlpolicy.CategoryConfig{
			urlpolicy.CategoryRunnerAPI:       {Entries: []string{"https://runner.example.com"}},
			urlpolicy.CategoryPublicAPI:       {Entries: []string{"https://api.example.com"}},
			urlpolicy.CategoryRunnerBootstrap: {Entries: []string{"https://raw.githubusercontent.com/acme/runner/"}},
			urlpolicy.CategoryStackTemplates:  {Entries: []string{"https://acme-templates.s3.us-west-2.amazonaws.com/"}},
		},
	})
	require.NoError(t, err)
	return policy
}

func violationFields(violations urlpolicy.Violations) []string {
	fields := make([]string, 0, len(violations))
	for _, violation := range violations {
		fields = append(fields, violation.Field)
	}
	return fields
}

func TestURLPolicyViolations(t *testing.T) {
	policy := testURLPolicy(t, urlpolicy.ModeEnforce)

	allowed := &config.AppConfig{
		Runner: &config.AppRunnerConfig{
			RunnerAPIURL:  "https://runner.example.com",
			PublicAPIURL:  "https://api.example.com",
			InitScriptURL: "https://raw.githubusercontent.com/acme/runner/main/init.sh#default",
			EnvVarMap:     map[string]string{"LOG_LEVEL": "debug"},
		},
		Stack: &config.StackConfig{
			Type:                 "aws-cloudformation",
			VPCNestedTemplateURL: "https://acme-templates.s3.us-west-2.amazonaws.com/vpc.yaml",
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "local", TemplateURL: "./stacks/local.yaml"},
				{Name: "remote", TemplateURL: "https://acme-templates.s3.us-west-2.amazonaws.com/extra.yaml"},
			},
		},
	}
	require.Empty(t, URLPolicyViolations(policy, allowed))

	denied := &config.AppConfig{
		Runner: &config.AppRunnerConfig{
			RunnerAPIURL:       "https://evil.test",
			PublicAPIURL:       "https://evil.test",
			InitScriptURL:      "https://raw.githubusercontent.com/mallory/runner/init.sh",
			PhoneHomeScriptURL: "git::https://github.com/mallory/x//phone_home.py",
			EnvVarMap:          map[string]string{"HTTPS_PROXY": "http://evil.test", "LOG_LEVEL": "debug"},
		},
		Stack: &config.StackConfig{
			Type:                    "aws-cloudformation",
			VPCNestedTemplateURL:    "https://mallory.s3.us-west-2.amazonaws.com/vpc.yaml",
			RunnerNestedTemplateURL: "https://mallory.s3.us-west-2.amazonaws.com/runner.yaml",
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "local", TemplateURL: "./stacks/local.yaml"},
				{Name: "remote", TemplateURL: "https://mallory.s3.us-west-2.amazonaws.com/extra.yaml"},
			},
		},
		Installs: []*config.Install{
			nil,
			{Name: "acme-prod", StackOverrides: &config.InstallStackOverrides{
				RunnerNestedTemplateURL: "https://mallory.s3.us-west-2.amazonaws.com/runner.yaml",
			}},
		},
	}
	require.Equal(t, []string{
		"runner.runner_api_url",
		"runner.public_api_url",
		"runner.init_script_url",
		"runner.phone_home_script_url",
		"runner.env_vars.HTTPS_PROXY",
		"stack.vpc_nested_template_url",
		"stack.runner_nested_template_url",
		"stack.custom_nested_stacks[1].template_url",
		"installs[acme-prod].stack_overrides.runner_nested_template_url",
	}, violationFields(URLPolicyViolations(policy, denied)))

	// An unrestricted policy accepts the same config.
	require.Empty(t, URLPolicyViolations(nil, denied))
}

func TestURLPolicyViolationsSkipsGCPCuratedModules(t *testing.T) {
	policy := testURLPolicy(t, urlpolicy.ModeEnforce)
	cfg := &config.AppConfig{Stack: &config.StackConfig{
		Type: "gcp-terraform",
		CustomNestedStacks: []config.CustomNestedStack{
			{Name: "bucket", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket"},
		},
	}}
	require.Empty(t, URLPolicyViolations(policy, cfg))
}

func TestValidateURLPolicyModes(t *testing.T) {
	cfg := &config.AppConfig{Runner: &config.AppRunnerConfig{RunnerAPIURL: "https://evil.test"}}

	err := ValidateURLPolicy(testURLPolicy(t, urlpolicy.ModeEnforce), cfg)
	require.Error(t, err)
	require.False(t, config.IsWarningErr(err))
	require.ErrorContains(t, err, "allowed_hosts_runner_api")

	err = ValidateURLPolicy(testURLPolicy(t, urlpolicy.ModeWarn), cfg)
	require.Error(t, err)
	require.True(t, config.IsWarningErr(err))

	require.NoError(t, ValidateURLPolicy(nil, cfg))
}
