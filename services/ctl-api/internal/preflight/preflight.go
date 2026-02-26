package preflight

import (
	"fmt"
	"strings"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

// Result holds the outcome of a single preflight check.
type Result struct {
	Name   string
	Passed bool
	Detail string
}

// Check is a function that validates connectivity or credentials for a config subset.
type Check func(cfg *internal.Config) (detail string, err error)

// Run executes the requested checks (or all if names is empty) and returns results.
func Run(cfg *internal.Config, names []string) []Result {
	if len(names) == 0 {
		names = ListChecks(cfg)
	}

	results := make([]Result, 0, len(names))
	for _, name := range names {
		check, ok := Checks[name]
		if !ok {
			results = append(results, Result{
				Name:   name,
				Passed: false,
				Detail: "unknown check",
			})
			continue
		}

		// Validate config fields for this check first.
		if _, err := ValidateConfigForCheck(cfg, name); err != nil {
			results = append(results, Result{
				Name:   name,
				Passed: false,
				Detail: err.Error(),
			})
			continue
		}

		detail, err := check(cfg)
		if err != nil {
			results = append(results, Result{
				Name:   name,
				Passed: false,
				Detail: err.Error(),
			})
		} else {
			results = append(results, Result{
				Name:   name,
				Passed: true,
				Detail: detail,
			})
		}
	}

	return results
}

// PrintResults prints a formatted table of check results and returns the exit code.
func PrintResults(results []Result) int {
	passed := 0
	for _, r := range results {
		passed += boolToInt(r.Passed)
	}

	fmt.Println()
	fmt.Printf("  %-14s %-10s %s\n", "CHECK", "STATUS", "DETAIL")

	for _, r := range results {
		status := "\033[32m✓ pass\033[0m"
		if !r.Passed {
			status = "\033[31m✗ fail\033[0m"
		}
		// Truncate long details for the table.
		detail := r.Detail
		if len(detail) > 80 {
			detail = detail[:77] + "..."
		}
		fmt.Printf("  %-14s %-10s %s\n", r.Name, status, detail)
	}

	fmt.Println()
	fmt.Printf("  %d/%d checks passed.", passed, len(results))
	if passed < len(results) {
		fmt.Println(" Exit code: 1")
		return 1
	}
	fmt.Println(" Exit code: 0")
	return 0
}

// FormatFieldSummary returns a parenthesized summary of key config values for a check.
func FormatFieldSummary(pairs ...string) string {
	if len(pairs) == 0 {
		return ""
	}
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		key, val := pairs[i], pairs[i+1]
		if val != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", key, val))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
