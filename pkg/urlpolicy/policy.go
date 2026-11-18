package urlpolicy

import (
	"fmt"
	"strings"
)

// Category groups app-config URLs by what the URL is trusted to do. The lists
// are separate so that allowing a template bucket does not also allow that
// bucket to receive the runner token.
type Category string

const (
	// CategoryRunnerAPI covers the API the runner sends its token to.
	CategoryRunnerAPI Category = "runner_api"
	// CategoryPublicAPI covers the API the install stack phones home to.
	CategoryPublicAPI Category = "public_api"
	// CategoryRunnerBootstrap covers scripts that run on the runner host or as
	// the phone-home function.
	CategoryRunnerBootstrap Category = "runner_bootstrap"
	// CategoryStackTemplates covers templates the customer's cloud deploys.
	CategoryStackTemplates Category = "stack_templates"
	// CategoryConfigFetch covers files ctl-api fetches while parsing a config.
	CategoryConfigFetch Category = "config_fetch"
)

// Categories lists every category in a stable order.
var Categories = []Category{
	CategoryRunnerAPI,
	CategoryPublicAPI,
	CategoryRunnerBootstrap,
	CategoryStackTemplates,
	CategoryConfigFetch,
}

// ConfigKey is the ctl-api config key an operator sets to restrict a category.
func (c Category) ConfigKey() string {
	return "allowed_hosts_" + string(c)
}

type Mode string

const (
	// ModeEnforce rejects a violation.
	ModeEnforce Mode = "enforce"
	// ModeWarn reports a violation without rejecting it, for an operator
	// trialling a new list.
	ModeWarn Mode = "warn"
)

// CategoryConfig is the input for one category.
type CategoryConfig struct {
	// Entries are the operator's entries. A category with none is not
	// restricted.
	Entries []string
	// Builtins are added once the category is restricted, so an operator does
	// not have to restate the origins a stock app config points at. They never
	// restrict a category on their own.
	Builtins []string
}

// Config is the input for New.
type Config struct {
	// Mode defaults to ModeEnforce.
	Mode       Mode
	AllowHTTP  bool
	Categories map[Category]CategoryConfig
}

// Policy is the effective, parsed policy. The zero value and a nil *Policy
// restrict nothing.
type Policy struct {
	mode       Mode
	categories map[Category][]Entry
	onWarn     func(Violations)
}

// New parses every entry and fails on the first malformed one.
func New(cfg Config) (*Policy, error) {
	mode := cfg.Mode
	switch mode {
	case "":
		mode = ModeEnforce
	case ModeEnforce, ModeWarn:
	default:
		return nil, fmt.Errorf("allowed_hosts_mode must be %q or %q, got %q", ModeEnforce, ModeWarn, mode)
	}

	opts := ParseOptions{AllowHTTP: cfg.AllowHTTP}
	policy := &Policy{mode: mode, categories: map[Category][]Entry{}}
	for category, categoryCfg := range cfg.Categories {
		if !knownCategory(category) {
			return nil, fmt.Errorf("unknown allowed hosts category %q", category)
		}

		var configured []string
		for _, raw := range categoryCfg.Entries {
			if strings.TrimSpace(raw) != "" {
				configured = append(configured, raw)
			}
		}
		if len(configured) == 0 {
			continue
		}

		seen := map[string]struct{}{}
		var entries []Entry
		for _, raw := range append(append([]string{}, categoryCfg.Builtins...), configured...) {
			entry, err := ParseEntry(raw, opts)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", category.ConfigKey(), err)
			}
			key := entry.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			entries = append(entries, entry)
		}
		policy.categories[category] = entries
	}

	return policy, nil
}

func knownCategory(category Category) bool {
	for _, known := range Categories {
		if known == category {
			return true
		}
	}
	return false
}

// Mode returns the policy's mode.
func (p *Policy) Mode() Mode {
	if p == nil || p.mode == "" {
		return ModeEnforce
	}
	return p.mode
}

// Restricted reports whether the category has an allowlist. An unrestricted
// category accepts any URL.
func (p *Policy) Restricted(category Category) bool {
	if p == nil {
		return false
	}
	return len(p.categories[category]) > 0
}

// AnyRestricted reports whether any category has an allowlist.
func (p *Policy) AnyRestricted() bool {
	if p == nil {
		return false
	}
	for _, entries := range p.categories {
		if len(entries) > 0 {
			return true
		}
	}
	return false
}

// Entries returns the effective entries for a category in canonical form:
// built-ins first, then the operator's. It is empty for an unrestricted
// category.
func (p *Policy) Entries(category Category) []string {
	if p == nil {
		return nil
	}
	out := make([]string, 0, len(p.categories[category]))
	for _, entry := range p.categories[category] {
		out = append(out, entry.String())
	}
	return out
}

// Check returns a violation when the category is restricted and rawURL does
// not fall under any of its entries. field names the config field, for the
// message. An empty rawURL is never a violation: the field is unset.
func (p *Policy) Check(category Category, field, rawURL string) *Violation {
	if !p.Restricted(category) || strings.TrimSpace(rawURL) == "" {
		return nil
	}

	c, reason := parseCandidate(rawURL)
	if reason != "" {
		return &Violation{Category: category, Field: field, Value: rawURL, Reason: reason}
	}
	for _, entry := range p.categories[category] {
		if entry.matches(c) {
			return nil
		}
	}
	return &Violation{
		Category: category,
		Field:    field,
		Value:    rawURL,
		Reason:   "it does not match any allowed entry",
	}
}

// OnWarn registers the function ReportWarnings calls. Set it once at startup,
// before the policy is shared.
func (p *Policy) OnWarn(fn func(Violations)) {
	if p != nil {
		p.onWarn = fn
	}
}

// ReportWarnings hands violations found in ModeWarn to the registered
// function.
func (p *Policy) ReportWarnings(violations Violations) {
	if p == nil || p.onWarn == nil || len(violations) == 0 {
		return
	}
	p.onWarn(violations)
}

// Violation is one config value the policy does not allow.
type Violation struct {
	Category Category `json:"category"`
	Field    string   `json:"field"`
	Value    string   `json:"value"`
	Reason   string   `json:"reason"`
}

func (v Violation) Error() string {
	return fmt.Sprintf(
		"%s %q is not allowed by this control plane: %s. The control plane operator controls this with the %s setting.",
		v.Field, v.Value, v.Reason, v.Category.ConfigKey(),
	)
}

// Violations is a list of violations that reads as one error.
type Violations []Violation

func (v Violations) Error() string {
	lines := make([]string, 0, len(v))
	for _, violation := range v {
		lines = append(lines, violation.Error())
	}
	return strings.Join(lines, "\n")
}

// reservedEnvVarNames are AWS SDK variables that redirect where credentials
// come from or where requests go.
var reservedEnvVarNames = map[string]struct{}{
	"AWS_ACCESS_KEY_ID":                   {},
	"AWS_SECRET_ACCESS_KEY":               {},
	"AWS_SESSION_TOKEN":                   {},
	"AWS_PROFILE":                         {},
	"AWS_DEFAULT_PROFILE":                 {},
	"AWS_ROLE_ARN":                        {},
	"AWS_ROLE_SESSION_NAME":               {},
	"AWS_WEB_IDENTITY_TOKEN_FILE":         {},
	"AWS_SHARED_CREDENTIALS_FILE":         {},
	"AWS_CONFIG_FILE":                     {},
	"AWS_CA_BUNDLE":                       {},
	"AWS_USE_FIPS_ENDPOINT":               {},
	"AWS_USE_DUALSTACK_ENDPOINT":          {},
	"AWS_IGNORE_CONFIGURED_ENDPOINT_URLS": {},
}

var reservedEnvVarPrefixes = []string{
	"RUNNER_",
	"NUON_",
	"AWS_ENDPOINT_URL",
	"AWS_CONTAINER_",
	"AWS_EC2_METADATA_",
}

// ReservedRunnerEnvVar reports whether a runner env var name could redirect
// the runner's traffic or credentials, or change which runner build runs.
func ReservedRunnerEnvVar(name string) bool {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if strings.HasSuffix(upper, "_PROXY") {
		return true
	}
	if _, ok := reservedEnvVarNames[upper]; ok {
		return true
	}
	for _, prefix := range reservedEnvVarPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// CheckRunnerEnvVar returns a violation for a reserved runner env var. The
// denylist applies only once the operator has restricted the runner API, which
// is how they opt into locking the runner down.
func (p *Policy) CheckRunnerEnvVar(field, name string) *Violation {
	if !p.Restricted(CategoryRunnerAPI) || !ReservedRunnerEnvVar(name) {
		return nil
	}
	return &Violation{
		Category: CategoryRunnerAPI,
		Field:    field,
		Value:    name,
		Reason:   "this runner environment variable is reserved while the runner API is restricted",
	}
}

// FromEffective rebuilds a policy from the effective entries a control plane
// serves, so a client matches URLs exactly as that control plane does. The
// entries already include built-ins. Categories this build does not know are
// skipped: it has no fields mapped to them.
func FromEffective(mode Mode, entries map[string][]string) (*Policy, error) {
	categories := make(map[Category]CategoryConfig, len(entries))
	for name, list := range entries {
		category := Category(name)
		if !knownCategory(category) {
			continue
		}
		categories[category] = CategoryConfig{Entries: list}
	}
	// The control plane already decided whether http entries are acceptable.
	return New(Config{Mode: mode, AllowHTTP: true, Categories: categories})
}
