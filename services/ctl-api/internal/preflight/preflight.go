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

type warning struct{ msg string }

func (w *warning) Error() string { return w.msg }

func warnf(format string, args ...any) error {
	return &warning{msg: fmt.Sprintf(format, args...)}
}

func isWarning(err error) bool {
	var w *warning

	return errors.As(err, &w)
}

type Field struct {
	Name     string
	Value    string
	Required bool
	Secret   bool
}

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

type Check struct {
	Name        string
	Description string

	Skip func(cfg *internal.Config) (string, bool)

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

func Describe(cfg *internal.Config, names []string) []Result {
	checks, unknown := resolve(names)

	results := make([]Result, 0, len(checks)+len(unknown))
	for _, name := range unknown {
		results = append(results, Result{Name: name, Status: StatusFail, Detail: "unknown check"})
	}

	for _, check := range checks {
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
