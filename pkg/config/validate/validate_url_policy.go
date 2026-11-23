package validate

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/urlpolicy"
)

// This file is the one place that says which config field belongs to which
// allowed-hosts category. The CLI and ctl-api both go through it.

type options struct {
	urlPolicy *urlpolicy.Policy
}

// Option configures Validate.
type Option func(*options)

// WithURLPolicy makes Validate check the config's URLs against the control
// plane's allowed hosts. A nil policy restricts nothing.
func WithURLPolicy(policy *urlpolicy.Policy) Option {
	return func(o *options) {
		o.urlPolicy = policy
	}
}

// RunnerURLs are the runner config values the policy governs.
type RunnerURLs struct {
	RunnerAPIURL       string
	PublicAPIURL       string
	InitScriptURL      string
	PhoneHomeScriptURL string
	EnvVars            map[string]string
}

// URLRef is one policy-governed URL found in a config: the value, the config
// path it came from and the category that governs it.
type URLRef struct {
	Category urlpolicy.Category
	Field    string
	Value    string
}

// RunnerURLRefs lists the URLs in the runner section of an app config. Unset
// values are left out.
func RunnerURLRefs(runner RunnerURLs) []URLRef {
	return nonEmptyRefs([]URLRef{
		{urlpolicy.CategoryRunnerAPI, "runner.runner_api_url", runner.RunnerAPIURL},
		{urlpolicy.CategoryPublicAPI, "runner.public_api_url", runner.PublicAPIURL},
		{urlpolicy.CategoryRunnerBootstrap, "runner.init_script_url", runner.InitScriptURL},
		{urlpolicy.CategoryRunnerBootstrap, "runner.phone_home_script_url", runner.PhoneHomeScriptURL},
	})
}

// RunnerURLPolicyViolations checks the runner section of an app config.
func RunnerURLPolicyViolations(policy *urlpolicy.Policy, runner RunnerURLs) urlpolicy.Violations {
	out := refViolations(policy, RunnerURLRefs(runner))

	names := make([]string, 0, len(runner.EnvVars))
	for name := range runner.EnvVars {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if v := policy.CheckRunnerEnvVar("runner.env_vars."+name, name); v != nil {
			out = append(out, *v)
		}
	}
	return out
}

// StackURLs are the stack template values the policy governs. The app stack
// config and the per-install stack overrides share this shape.
type StackURLs struct {
	// Type is the app's stack type. Custom nested stacks on GCP name curated
	// modules, not URLs, and are validated against that list instead.
	Type                    string
	VPCNestedTemplateURL    string
	RunnerNestedTemplateURL string
	UIDefinition            string
	CustomNestedStacks      []config.CustomNestedStack
}

// StackURLRefs lists the remote stack template URLs in a stack section. field
// is the config path of the section, "stack" for the app stack config. Unset
// values and files in the config repo are left out.
func StackURLRefs(field string, stack StackURLs) []URLRef {
	refs := []URLRef{
		{urlpolicy.CategoryStackTemplates, field + ".vpc_nested_template_url", stack.VPCNestedTemplateURL},
		{urlpolicy.CategoryStackTemplates, field + ".runner_nested_template_url", stack.RunnerNestedTemplateURL},
	}
	if !isLocalPath(stack.UIDefinition) {
		refs = append(refs, URLRef{urlpolicy.CategoryStackTemplates, field + ".ui_definition", stack.UIDefinition})
	}
	if stack.Type != "gcp-terraform" {
		for i, nested := range stack.CustomNestedStacks {
			if isLocalPath(nested.TemplateURL) {
				continue
			}
			refs = append(refs, URLRef{
				urlpolicy.CategoryStackTemplates,
				fmt.Sprintf("%s.custom_nested_stacks[%d].template_url", field, i),
				nested.TemplateURL,
			})
		}
	}
	return nonEmptyRefs(refs)
}

// StackURLPolicyViolations checks stack template URLs. field is the config
// path of the section, "stack" for the app stack config.
func StackURLPolicyViolations(policy *urlpolicy.Policy, field string, stack StackURLs) urlpolicy.Violations {
	return refViolations(policy, StackURLRefs(field, stack))
}

func nonEmptyRefs(refs []URLRef) []URLRef {
	out := refs[:0]
	for _, ref := range refs {
		if strings.TrimSpace(ref.Value) != "" {
			out = append(out, ref)
		}
	}
	return out
}

func refViolations(policy *urlpolicy.Policy, refs []URLRef) urlpolicy.Violations {
	var out urlpolicy.Violations
	for _, ref := range refs {
		if v := policy.Check(ref.Category, ref.Field, ref.Value); v != nil {
			out = append(out, *v)
		}
	}
	return out
}

// isLocalPath reports whether a template reference is a file in the config
// repo. Its contents are read at parse time and re-hosted, so there is no
// remote location to restrict.
func isLocalPath(value string) bool {
	if strings.HasPrefix(value, "//") {
		return false
	}
	return strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") || strings.HasPrefix(value, "/")
}

// URLPolicyViolations checks every policy-governed value in an app config.
func URLPolicyViolations(policy *urlpolicy.Policy, a *config.AppConfig) urlpolicy.Violations {
	if a == nil || !policy.AnyRestricted() {
		return nil
	}

	var out urlpolicy.Violations
	if a.Runner != nil {
		out = append(out, RunnerURLPolicyViolations(policy, RunnerURLs{
			RunnerAPIURL:       a.Runner.RunnerAPIURL,
			PublicAPIURL:       a.Runner.PublicAPIURL,
			InitScriptURL:      a.Runner.InitScriptURL,
			PhoneHomeScriptURL: a.Runner.PhoneHomeScriptURL,
			EnvVars:            a.Runner.EnvVarMap,
		})...)
	}

	stackType := ""
	if a.Stack != nil {
		stackType = a.Stack.Type
		out = append(out, StackURLPolicyViolations(policy, "stack", StackURLs{
			Type:                    stackType,
			VPCNestedTemplateURL:    a.Stack.VPCNestedTemplateURL,
			RunnerNestedTemplateURL: a.Stack.RunnerNestedTemplateURL,
			UIDefinition:            a.Stack.UIDefinition,
			CustomNestedStacks:      a.Stack.CustomNestedStacks,
		})...)
	}

	for _, install := range a.Installs {
		if install == nil || install.StackOverrides == nil {
			continue
		}
		overrides := install.StackOverrides
		out = append(out, StackURLPolicyViolations(policy, fmt.Sprintf("installs[%s].stack_overrides", install.Name), StackURLs{
			Type:                    stackType,
			VPCNestedTemplateURL:    overrides.VPCNestedTemplateURL,
			RunnerNestedTemplateURL: overrides.RunnerNestedTemplateURL,
			UIDefinition:            overrides.UIDefinition,
			CustomNestedStacks:      overrides.CustomNestedStacks,
		})...)
	}
	return out
}

// ValidateURLPolicy returns the config's violations as one error. In warn mode
// the error is a warning, which Validate reports without failing.
func ValidateURLPolicy(policy *urlpolicy.Policy, a *config.AppConfig) error {
	violations := URLPolicyViolations(policy, a)
	if len(violations) == 0 {
		return nil
	}
	return config.ErrConfig{
		Description: violations.Error(),
		Err:         violations,
		Warning:     policy.Mode() == urlpolicy.ModeWarn,
	}
}

// stackTemplateRedirectCheck is the check validation runs on each redirect
// target while it fetches a stack template. In enforce mode an off-list target
// fails the fetch. In warn mode it is added to warnings, which Validate
// reports with its other warnings, and the redirect is followed.
func stackTemplateRedirectCheck(policy *urlpolicy.Policy, warnings *[]string) redirectCheck {
	if !policy.Restricted(urlpolicy.CategoryStackTemplates) {
		return nil
	}
	return func(target *url.URL) error {
		violation := policy.Check(urlpolicy.CategoryStackTemplates, "redirect followed for a stack template", target.String())
		if violation == nil {
			return nil
		}
		if policy.Mode() == urlpolicy.ModeWarn {
			*warnings = append(*warnings, violation.Error())
			return nil
		}
		return *violation
	}
}

// stackTemplateFetchFields are the fetched fields whose source is a stack
// template. Every other fetched field is a plain config file.
var stackTemplateFetchFields = map[string]struct{}{
	"CustomNestedStack.Contents":                 {},
	"StackConfig.UIDefinitionContents":           {},
	"InstallStackOverrides.UIDefinitionContents": {},
}

// RemoteFetchCheck returns the parse.ParseConfig.CheckRemoteFetch hook for a
// policy: it rejects a remote file outside the allowed hosts before it is
// fetched. In warn mode the violation goes to onWarn and the fetch goes ahead.
// It returns nil when the policy restricts no fetch.
func RemoteFetchCheck(policy *urlpolicy.Policy, onWarn func(urlpolicy.Violation)) func(field, rawURL string) error {
	if !policy.Restricted(urlpolicy.CategoryConfigFetch) && !policy.Restricted(urlpolicy.CategoryStackTemplates) {
		return nil
	}
	return func(field, rawURL string) error {
		category := urlpolicy.CategoryConfigFetch
		if _, ok := stackTemplateFetchFields[field]; ok {
			category = urlpolicy.CategoryStackTemplates
		}
		violation := policy.Check(category, "file fetched for "+field, rawURL)
		if violation == nil {
			return nil
		}
		if policy.Mode() == urlpolicy.ModeWarn {
			if onWarn != nil {
				onWarn(*violation)
			}
			return nil
		}
		return config.ErrConfig{Description: violation.Error(), Err: *violation}
	}
}
