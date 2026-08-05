package preflight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusWarn    Status = "warn"
	StatusFail    Status = "fail"
	StatusSkipped Status = "skipped"
)

// warning is returned by a probe that validated everything the config allows
// but could not confirm it against the live service — a provider with fixed
// endpoints and nothing to discover, say. Reported distinctly from a pass so
// the gap is visible, but not counted as a failure.
type warning struct{ msg string }

func (w *warning) Error() string { return w.msg }

func warnf(format string, args ...any) error {
	return &warning{msg: fmt.Sprintf(format, args...)}
}

func isWarning(err error) bool {
	var w *warning

	return errors.As(err, &w)
}

// Field is one config value a check depends on. Value is pre-rendered by the
// check so typed fields (bool, time.Duration) format themselves.
type Field struct {
	Name     string
	Value    string
	Required bool
	Secret   bool
}

// Display renders a field value for output, masking secrets.
func (f Field) Display() string {
	switch {
	case f.Value == "":
		return "(unset)"
	case f.Secret:
		return "******"
	default:
		return f.Value
	}
}

// Check declares an external dependency, the config it reads, and how to probe it.
type Check struct {
	Name        string
	Description string

	// Skip reports why the check does not apply to this config, e.g. a
	// cloud_provider or feature flag that turns the dependency off.
	Skip func(cfg *internal.Config) (string, bool)

	// Fields is a func rather than a static list so required-ness can depend on
	// other config: db_password is only required when db_use_iam is off.
	Fields func(cfg *internal.Config) []Field

	Probe func(ctx context.Context, cfg *internal.Config) (string, error)
}

type Result struct {
	Name        string
	Description string
	Status      Status
	Detail      string
	Fields      []Field
}

func (r Result) Failed() bool { return r.Status == StatusFail }

// Run executes the named checks, or every check in registry order when names is
// empty. A check is skipped, then field-validated, then probed.
func Run(ctx context.Context, cfg *internal.Config, names []string) []Result {
	checks, unknown := resolve(names)

	results := make([]Result, 0, len(checks)+len(unknown))
	for _, name := range unknown {
		results = append(results, Result{
			Name:   name,
			Status: StatusFail,
			Detail: "unknown check",
		})
	}

	for _, check := range checks {
		results = append(results, run(ctx, cfg, check))
	}

	return results
}

// Describe evaluates skip state and field values for the named checks without
// probing anything, backing `preflight --list`.
func Describe(cfg *internal.Config, names []string) []Result {
	checks, unknown := resolve(names)

	results := make([]Result, 0, len(checks)+len(unknown))
	for _, name := range unknown {
		results = append(results, Result{Name: name, Status: StatusFail, Detail: "unknown check"})
	}

	for _, check := range checks {
		// No Status: nothing was run, and a listing that claimed otherwise
		// would be misleading.
		result := Result{
			Name:        check.Name,
			Description: check.Description,
			Fields:      check.Fields(cfg),
		}
		if check.Skip != nil {
			if reason, skip := check.Skip(cfg); skip {
				result.Status = StatusSkipped
				result.Detail = reason
			}
		}
		results = append(results, result)
	}

	return results
}

func run(ctx context.Context, cfg *internal.Config, check Check) Result {
	// Fields are resolved even when the check is skipped, so a run and a --list
	// describe a check identically. Consumers generating docs from the JSON get
	// the same field set either way.
	result := Result{
		Name:        check.Name,
		Description: check.Description,
		Fields:      check.Fields(cfg),
	}

	if check.Skip != nil {
		if reason, skip := check.Skip(cfg); skip {
			result.Status = StatusSkipped
			result.Detail = reason

			return result
		}
	}

	if missing := missingFields(result.Fields); len(missing) > 0 {
		result.Status = StatusFail
		result.Detail = "missing required config: " + strings.Join(missing, ", ")

		return result
	}

	// Bounded per check so one black-holed host cannot stall the whole run.
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	detail, err := check.Probe(ctx, cfg)
	switch {
	case err == nil:
		result.Status = StatusPass
		result.Detail = detail
	case isWarning(err):
		result.Status = StatusWarn
		result.Detail = err.Error()
	default:
		result.Status = StatusFail
		result.Detail = err.Error()
	}

	return result
}

func missingFields(fields []Field) []string {
	var missing []string
	for _, f := range fields {
		if f.Required && f.Value == "" {
			missing = append(missing, f.Name)
		}
	}

	return missing
}

// resolve maps requested names onto registry entries, preserving registry order
// so output is stable, and reports any name with no matching check.
func resolve(names []string) ([]Check, []string) {
	if len(names) == 0 {
		return All(), nil
	}

	wanted := make(map[string]bool, len(names))
	var unknown []string
	for _, name := range names {
		if _, ok := Lookup(name); !ok {
			unknown = append(unknown, name)
			continue
		}
		wanted[name] = true
	}

	var checks []Check
	for _, check := range All() {
		if wanted[check.Name] {
			checks = append(checks, check)
		}
	}

	return checks, unknown
}

// summary renders a short "(key=value, ...)" tail for a probe detail line.
func summary(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", pairs[i], pairs[i+1]))
		}
	}
	if len(parts) == 0 {
		return ""
	}

	return "(" + strings.Join(parts, ", ") + ")"
}
