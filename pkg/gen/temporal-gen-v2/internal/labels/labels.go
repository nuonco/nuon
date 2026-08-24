// Package labels loads the temporal-gen.yaml config file, which declares label
// keys, their permitted values, and the default option attributes each value
// implies for activities and workflows.
//
// Labels are key/value pairs set on a function with `@label <key> <value>`,
// mirroring the existing `@memo <key> <value>` form. A label's attributes are
// lowered back into synthetic annotation comment lines (see AnnotationLines)
// and fed through the normal annotation parser. That way there is exactly one
// code path turning "@start-to-close-timeout 30s" into a field assignment, and
// label defaults get last-write-wins precedence for free by being emitted
// ahead of a function's own annotations.
package labels

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SupportedVersion is the only value accepted for the top-level `version` key.
const SupportedVersion = 1

// FileNames are the config file names searched for, in order, at each level of
// the upward walk performed by Discover.
var FileNames = []string{"temporal-gen.yaml", "temporal-gen.yml"}

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// Pair is a single `@label <key> <value>` set on a function.
type Pair struct {
	Key   string
	Value string
}

func (p Pair) String() string { return p.Key + "=" + p.Value }

// Config is a parsed temporal-gen.yaml.
type Config struct {
	Version  int             `yaml:"version"`
	Defaults *Attrs          `yaml:"defaults"`
	Labels   map[string]*Key `yaml:"labels"`

	path string
}

// Key is a label key together with the values it permits.
type Key struct {
	Description string            `yaml:"description"`
	Values      map[string]*Value `yaml:"values"`
}

// Value is one permitted value of a label key, and the defaults it implies.
type Value struct {
	Description string `yaml:"description"`
	Attrs       `yaml:",inline"`
}

// Attrs holds the per-kind default blocks.
type Attrs struct {
	Activity *ActivityAttrs `yaml:"activity"`
	Workflow *WorkflowAttrs `yaml:"workflow"`
}

// ActivityAttrs is the allowlist of activity annotations settable from a label.
//
// Structural annotations (@as-wrapper, @by-field, @local, @namespace,
// @options-callback, ...) are deliberately absent: they describe an individual
// function, not a class of them. Booleans are also one-way in the annotation
// language (there is no "off" form), so a label that could set @as-wrapper
// could never be opted out of by a function.
type ActivityAttrs struct {
	TaskQueue              string    `yaml:"task-queue"`
	ScheduleToCloseTimeout *duration `yaml:"schedule-to-close-timeout"`
	ScheduleToStartTimeout *duration `yaml:"schedule-to-start-timeout"`
	StartToCloseTimeout    *duration `yaml:"start-to-close-timeout"`
	HeartbeatTimeout       *duration `yaml:"heartbeat-timeout"`
	WaitForCancellation    *bool     `yaml:"wait-for-cancellation"`
	DisableEagerExecution  *bool     `yaml:"disable-eager-execution"`
	MaxRetries             *int      `yaml:"max-retries"`
	RetryPolicyMaxAttempts *int      `yaml:"retry-policy-max-attempts"`
}

// WorkflowAttrs is the allowlist of workflow annotations settable from a label.
type WorkflowAttrs struct {
	TaskQueue           string            `yaml:"task-queue"`
	ExecutionTimeout    *duration         `yaml:"execution-timeout"`
	TaskTimeout         *duration         `yaml:"task-timeout"`
	WaitForCancellation *bool             `yaml:"wait-for-cancellation"`
	Memo                map[string]string `yaml:"memo"`
}

// duration is a time.Duration written in YAML as a Go duration string ("30s",
// "1h"). yaml.v3 has no native time.Duration support, and keeping the raw text
// means the emitted annotation line reads exactly as authored.
type duration struct {
	raw string
}

func (d *duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return fmt.Errorf("line %d: duration must be a quoted string such as \"30s\" or \"1h\": %w", node.Line, err)
	}
	if _, err := time.ParseDuration(s); err != nil {
		return fmt.Errorf("line %d: invalid duration %q: %w", node.Line, s, err)
	}
	d.raw = s
	return nil
}

// Path returns the file this config was loaded from.
func (c *Config) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

// Load reads and parses a config file. Unknown keys are a hard error so that a
// misspelled attribute fails loudly instead of silently doing nothing.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config %s: %w", path, err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	cfg.path = path

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Discover walks up from dir looking for a config file, stopping after the
// directory containing go.mod (the module root). It returns (nil, nil) when no
// config exists, which is the normal case for packages that do not use labels.
func Discover(dir string) (*Config, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", dir, err)
	}

	for {
		for _, name := range FileNames {
			p := filepath.Join(abs, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return Load(p)
			}
		}

		// Check the module root only after checking this directory, so a
		// config sitting next to go.mod is still found.
		if st, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil && !st.IsDir() {
			return nil, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			return nil, nil
		}
		abs = parent
	}
}

// Validate checks structural invariants that YAML decoding cannot express.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	where := c.path
	if where == "" {
		where = "config"
	}

	if c.Version == 0 {
		return fmt.Errorf("%s: `version` is required (expected %d)", where, SupportedVersion)
	}
	if c.Version != SupportedVersion {
		return fmt.Errorf("%s: unsupported version %d (expected %d)", where, c.Version, SupportedVersion)
	}

	if c.Defaults != nil && c.Defaults.empty() {
		return fmt.Errorf("%s: `defaults` defines no activity or workflow attributes", where)
	}

	for _, key := range c.Keys() {
		k := c.Labels[key]
		if !nameRe.MatchString(key) {
			return fmt.Errorf("%s: invalid label key %q (must match %s)", where, key, nameRe)
		}
		if k == nil || len(k.Values) == 0 {
			return fmt.Errorf("%s: label key %q declares no values", where, key)
		}
		for _, value := range c.Values(key) {
			v := k.Values[value]
			if !nameRe.MatchString(value) {
				return fmt.Errorf("%s: invalid value %q for label key %q (must match %s)", where, value, key, nameRe)
			}
			if v == nil || v.empty() {
				return fmt.Errorf("%s: label %s=%s defines no activity or workflow attributes", where, key, value)
			}
		}
	}

	return nil
}

// Keys returns the declared label keys in sorted order.
func (c *Config) Keys() []string {
	if c == nil {
		return nil
	}
	keys := make([]string, 0, len(c.Labels))
	for k := range c.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Values returns the declared values for a label key, in sorted order.
func (c *Config) Values(key string) []string {
	if c == nil || c.Labels[key] == nil {
		return nil
	}
	values := make([]string, 0, len(c.Labels[key].Values))
	for v := range c.Labels[key].Values {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

// AnnotationLines lowers the defaults block plus each label pair into
// annotation comment lines for the given kind ("activity" or "workflow").
//
// Lines are emitted defaults-first, then pairs in the order the caller listed
// them (source order), so a later @label overrides an earlier one. A label
// whose value carries no block for kind contributes nothing: that is what lets
// one label serve both activities and workflows.
func (c *Config) AnnotationLines(kind string, pairs []Pair) ([]string, error) {
	if c == nil {
		if len(pairs) > 0 {
			return nil, fmt.Errorf("@label %s used but no %s was found", pairs[0], FileNames[0])
		}
		return nil, nil
	}

	seen := make(map[string]string, len(pairs))
	var lines []string

	if c.Defaults != nil {
		lines = append(lines, c.Defaults.lines(kind)...)
	}

	for _, p := range pairs {
		if prev, ok := seen[p.Key]; ok {
			return nil, fmt.Errorf("label key %q set twice (%q then %q); a key may only have one value", p.Key, prev, p.Value)
		}
		seen[p.Key] = p.Value

		k, ok := c.Labels[p.Key]
		if !ok {
			return nil, fmt.Errorf("unknown label key %q (declared in %s: %s)", p.Key, c.path, joinOrNone(c.Keys()))
		}
		v, ok := k.Values[p.Value]
		if !ok {
			return nil, fmt.Errorf("unknown value %q for label key %q (declared in %s: %s)", p.Value, p.Key, c.path, joinOrNone(c.Values(p.Key)))
		}
		lines = append(lines, v.lines(kind)...)
	}

	return lines, nil
}

func joinOrNone(s []string) string {
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ", ")
}

func (a *Attrs) empty() bool {
	return a == nil || (a.Activity == nil && a.Workflow == nil)
}

func (a *Attrs) lines(kind string) []string {
	if a == nil {
		return nil
	}
	switch kind {
	case "activity":
		return a.Activity.lines()
	case "workflow":
		return a.Workflow.lines()
	default:
		// Queries, signals and updates have no configurable options today.
		return nil
	}
}

func (a *ActivityAttrs) lines() []string {
	if a == nil {
		return nil
	}
	var out []string
	add := func(name, value string) {
		out = append(out, "// @"+name+" "+value)
	}

	if a.TaskQueue != "" {
		add("task-queue", a.TaskQueue)
	}
	if a.ScheduleToCloseTimeout != nil {
		add("schedule-to-close-timeout", a.ScheduleToCloseTimeout.raw)
	}
	if a.ScheduleToStartTimeout != nil {
		add("schedule-to-start-timeout", a.ScheduleToStartTimeout.raw)
	}
	if a.StartToCloseTimeout != nil {
		add("start-to-close-timeout", a.StartToCloseTimeout.raw)
	}
	if a.HeartbeatTimeout != nil {
		add("heartbeat-timeout", a.HeartbeatTimeout.raw)
	}
	if a.WaitForCancellation != nil {
		add("wait-for-cancellation", strconv.FormatBool(*a.WaitForCancellation))
	}
	if a.DisableEagerExecution != nil {
		add("disable-eager-execution", strconv.FormatBool(*a.DisableEagerExecution))
	}
	if a.MaxRetries != nil {
		add("max-retries", strconv.Itoa(*a.MaxRetries))
	}
	if a.RetryPolicyMaxAttempts != nil {
		add("retry-policy-max-attempts", strconv.Itoa(*a.RetryPolicyMaxAttempts))
	}

	return out
}

func (w *WorkflowAttrs) lines() []string {
	if w == nil {
		return nil
	}
	var out []string
	add := func(name, value string) {
		out = append(out, "// @"+name+" "+value)
	}

	if w.TaskQueue != "" {
		// Workflows read their task queue from @workflow-task-queue; plain
		// @task-queue is activity-only in the parser.
		add("workflow-task-queue", w.TaskQueue)
	}
	if w.ExecutionTimeout != nil {
		add("execution-timeout", w.ExecutionTimeout.raw)
	}
	if w.TaskTimeout != nil {
		add("task-timeout", w.TaskTimeout.raw)
	}
	if w.WaitForCancellation != nil {
		add("wait-for-cancellation", strconv.FormatBool(*w.WaitForCancellation))
	}

	// Sorted so generated output is deterministic across runs.
	keys := make([]string, 0, len(w.Memo))
	for k := range w.Memo {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		add("memo", k+" "+w.Memo[k])
	}

	return out
}
