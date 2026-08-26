// Package tags declares the tag vocabulary for temporal-gen-v2: the set of
// tag names an activity may carry, and the default activity options each name
// implies.
//
// Tags are single names set on an activity with `@tag <name>`. A config is
// either loaded from a temporal-gen.yaml or built in Go and handed to
// temporalgen.Options.Tags. Tags apply to activities only.
//
// A tag's attributes are lowered back into synthetic annotation comment lines
// (see AnnotationLines) and fed through the normal annotation parser. That way
// there is exactly one code path turning "@start-to-close-timeout 30s" into a
// field assignment, and tag defaults get last-write-wins precedence for free by
// being emitted ahead of a function's own annotations.
package tags

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

// Config is a parsed temporal-gen.yaml, or a vocabulary declared in Go.
type Config struct {
	Version  int               `yaml:"version"`
	Defaults *Attrs            `yaml:"defaults"`
	Tags     map[string]*Attrs `yaml:"tags"`

	path string
}

// Duration is a time.Duration written as a Go duration string ("30s", "1h").
// Keeping the raw text means the emitted annotation line reads exactly as
// authored, and an empty value means "unset".
type Duration string

// Attrs is the allowlist of activity annotations settable from a tag.
//
// Structural annotations (@as-wrapper, @by-field, @local, @namespace,
// @options-callback, ...) are deliberately absent: they describe an individual
// function, not a class of them. Booleans are also one-way in the annotation
// language (there is no "off" form), so a tag that could set @as-wrapper could
// never be opted out of by a function.
type Attrs struct {
	TaskQueue              string   `yaml:"task-queue"`
	ScheduleToCloseTimeout Duration `yaml:"schedule-to-close-timeout"`
	ScheduleToStartTimeout Duration `yaml:"schedule-to-start-timeout"`
	StartToCloseTimeout    Duration `yaml:"start-to-close-timeout"`
	HeartbeatTimeout       Duration `yaml:"heartbeat-timeout"`
	WaitForCancellation    *bool    `yaml:"wait-for-cancellation"`
	DisableEagerExecution  *bool    `yaml:"disable-eager-execution"`
	MaxRetries             *int     `yaml:"max-retries"`
	RetryPolicyMaxAttempts *int     `yaml:"retry-policy-max-attempts"`
}

// Path returns the file this config was loaded from, or "" when it was built
// in code.
func (c *Config) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

// source describes where a config came from, for error messages.
func (c *Config) source() string {
	if c == nil || c.path == "" {
		return "code"
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
// config exists, which is the normal case for packages that do not use tags.
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

// Validate checks structural invariants that YAML decoding cannot express. A
// config built in code may omit `version`; a file must declare it, so that a
// future format change has something to key off.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	where := c.source()

	if c.path == "" && c.Version == 0 {
		c.Version = SupportedVersion
	}
	if c.Version == 0 {
		return fmt.Errorf("%s: `version` is required (expected %d)", where, SupportedVersion)
	}
	if c.Version != SupportedVersion {
		return fmt.Errorf("%s: unsupported version %d (expected %d)", where, c.Version, SupportedVersion)
	}

	if c.Defaults != nil {
		if err := c.Defaults.validate(); err != nil {
			return fmt.Errorf("%s: `defaults`: %w", where, err)
		}
		if c.Defaults.empty() {
			return fmt.Errorf("%s: `defaults` sets no attributes", where)
		}
	}

	for _, name := range c.Names() {
		t := c.Tags[name]
		if !nameRe.MatchString(name) {
			return fmt.Errorf("%s: invalid tag %q (must match %s)", where, name, nameRe)
		}
		if t == nil || t.empty() {
			return fmt.Errorf("%s: tag %q sets no attributes", where, name)
		}
		if err := t.validate(); err != nil {
			return fmt.Errorf("%s: tag %q: %w", where, name, err)
		}
	}

	return nil
}

// Names returns the declared tags in sorted order.
func (c *Config) Names() []string {
	if c == nil {
		return nil
	}
	names := make([]string, 0, len(c.Tags))
	for name := range c.Tags {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AnnotationLines lowers the defaults block plus each tag into activity
// annotation comment lines.
//
// Lines are emitted defaults-first, then tags in the order the caller listed
// them (source order), so a later @tag overrides an earlier one.
func (c *Config) AnnotationLines(names []string) ([]string, error) {
	if c == nil {
		if len(names) > 0 {
			return nil, fmt.Errorf("@tag %s used but no tag config was found (add a %s, or set Options.Tags)", names[0], FileNames[0])
		}
		return nil, nil
	}

	seen := make(map[string]struct{}, len(names))
	lines := c.Defaults.lines()

	for _, name := range names {
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("tag %q set twice", name)
		}
		seen[name] = struct{}{}

		t, ok := c.Tags[name]
		if !ok {
			return nil, fmt.Errorf("unknown tag %q (declared in %s: %s)", name, c.source(), joinOrNone(c.Names()))
		}
		lines = append(lines, t.lines()...)
	}

	return lines, nil
}

func joinOrNone(s []string) string {
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ", ")
}

func (a *Attrs) empty() bool { return len(a.lines()) == 0 }

func (a *Attrs) validate() error {
	if a == nil {
		return nil
	}
	for name, d := range map[string]Duration{
		"schedule-to-close-timeout": a.ScheduleToCloseTimeout,
		"schedule-to-start-timeout": a.ScheduleToStartTimeout,
		"start-to-close-timeout":    a.StartToCloseTimeout,
		"heartbeat-timeout":         a.HeartbeatTimeout,
	} {
		if d == "" {
			continue
		}
		if _, err := time.ParseDuration(string(d)); err != nil {
			return fmt.Errorf("invalid duration %q for %s: %w", string(d), name, err)
		}
	}
	return nil
}

func (a *Attrs) lines() []string {
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
	if a.ScheduleToCloseTimeout != "" {
		add("schedule-to-close-timeout", string(a.ScheduleToCloseTimeout))
	}
	if a.ScheduleToStartTimeout != "" {
		add("schedule-to-start-timeout", string(a.ScheduleToStartTimeout))
	}
	if a.StartToCloseTimeout != "" {
		add("start-to-close-timeout", string(a.StartToCloseTimeout))
	}
	if a.HeartbeatTimeout != "" {
		add("heartbeat-timeout", string(a.HeartbeatTimeout))
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
