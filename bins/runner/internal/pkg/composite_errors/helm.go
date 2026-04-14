package composite_errors

import (
	"regexp"
	"strings"
)

var (
	// templateErrorRe matches Helm template errors like:
	//   Error: template: mychart/templates/deployment.yaml:12:5: executing "mychart/templates/deployment.yaml" at <.Values.missing>: ...
	templateErrorRe = regexp.MustCompile(`^Error:\s+template:\s+(\S+):(\d+):(\d+):\s+(.+)$`)

	// installFailedRe matches Helm install failures.
	installFailedRe = regexp.MustCompile(`^Error:\s+INSTALLATION FAILED:\s+(.+)$`)

	// upgradeFailedRe matches Helm upgrade failures.
	upgradeFailedRe = regexp.MustCompile(`^Error:\s+UPGRADE FAILED:\s+(.+)$`)

	// genericErrorRe matches any line starting with "Error:".
	genericErrorRe = regexp.MustCompile(`^Error:\s+(.+)$`)
)

// ParseHelmStderr parses helm stderr output into structured errors.
// Helm has no JSON error mode, so we parse known error patterns from stderr text.
func ParseHelmStderr(stderr string, ownerType string) []CompositeError {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return nil
	}

	var errors []CompositeError
	lines := strings.Split(stderr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if m := templateErrorRe.FindStringSubmatch(line); m != nil {
			errors = append(errors, CompositeError{
				OwnerType: ownerType,
				Severity:     "critical",
				Summary:      m[4],
				Detail:       line,
				Metadata: map[string]any{
					"template": m[1],
					"line":     m[2],
					"column":   m[3],
				},
			})
			continue
		}

		if m := installFailedRe.FindStringSubmatch(line); m != nil {
			errors = append(errors, CompositeError{
				OwnerType: ownerType,
				Severity:     "critical",
				Summary:      m[1],
				Detail:       line,
			})
			continue
		}

		if m := upgradeFailedRe.FindStringSubmatch(line); m != nil {
			errors = append(errors, CompositeError{
				OwnerType: ownerType,
				Severity:     "critical",
				Summary:      m[1],
				Detail:       line,
			})
			continue
		}

		if m := genericErrorRe.FindStringSubmatch(line); m != nil {
			errors = append(errors, CompositeError{
				OwnerType: ownerType,
				Severity:     "critical",
				Summary:      m[1],
				Detail:       line,
			})
			continue
		}
	}

	// If no recognized patterns were found, use the last non-empty line as summary
	// with the full stderr as detail.
	if len(errors) == 0 {
		lastLine := lastNonEmptyLine(lines)
		if lastLine != "" {
			errors = append(errors, CompositeError{
				OwnerType: ownerType,
				Severity:     "critical",
				Summary:      lastLine,
				Detail:       stderr,
			})
		}
	}

	return errors
}

func lastNonEmptyLine(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}
