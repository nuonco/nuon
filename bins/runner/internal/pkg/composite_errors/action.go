package composite_errors

import (
	"strings"
)

// ParseActionOutput parses action command output when a command fails.
// It uses the last non-empty line as the summary, and the last 20 lines as the detail.
func ParseActionOutput(output string, ownerType string) []CompositeError {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")

	summary := lastNonEmptyLine(lines)
	if summary == "" {
		return nil
	}

	detail := lastNLines(lines, 20)

	return []CompositeError{
		{
			OwnerType: ownerType,
			Severity:     "critical",
			Summary:      summary,
			Detail:       detail,
		},
	}
}

// lastNLines returns the last n lines joined as a single string.
// If there are fewer than n lines, all lines are returned.
func lastNLines(lines []string, n int) string {
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	return strings.Join(lines[start:], "\n")
}
