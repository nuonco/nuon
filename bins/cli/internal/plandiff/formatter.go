package plandiff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	colorGreen  = color.New(color.FgGreen)
	colorYellow = color.New(color.FgYellow)
	colorRed    = color.New(color.FgRed)
	colorCyan   = color.New(color.FgCyan)
	colorGray   = color.New(color.FgHiBlack)
	colorBold   = color.New(color.Bold)
)

func FormatSummary(summary Summary) string {
	var parts []string

	if summary.Create > 0 {
		parts = append(parts, colorGreen.Sprintf("%d to create", summary.Create))
	}
	if summary.Add > 0 {
		parts = append(parts, colorGreen.Sprintf("%d to add", summary.Add))
	}
	if summary.Update > 0 {
		parts = append(parts, colorYellow.Sprintf("%d to update", summary.Update))
	}
	if summary.Change > 0 {
		parts = append(parts, colorYellow.Sprintf("%d to change", summary.Change))
	}
	if summary.Replace > 0 {
		parts = append(parts, colorYellow.Sprintf("%d to replace", summary.Replace))
	}
	if summary.Delete > 0 {
		parts = append(parts, colorRed.Sprintf("%d to delete", summary.Delete))
	}
	if summary.Destroy > 0 {
		parts = append(parts, colorRed.Sprintf("%d to destroy", summary.Destroy))
	}
	if summary.Read > 0 {
		parts = append(parts, colorCyan.Sprintf("%d to read", summary.Read))
	}
	if summary.NoOp > 0 {
		parts = append(parts, colorGray.Sprintf("%d unchanged", summary.NoOp))
	}

	if len(parts) == 0 {
		return "No changes"
	}

	return "Plan: " + strings.Join(parts, ", ")
}

func ColorizeAction(action string) string {
	switch action {
	case "create", "add", "added":
		return colorGreen.Sprint(action)
	case "update", "change", "changed", "replace":
		return colorYellow.Sprint(action)
	case "delete", "destroy", "destroyed":
		return colorRed.Sprint(action)
	case "read":
		return colorCyan.Sprint(action)
	case "no-op":
		return colorGray.Sprint(action)
	default:
		return action
	}
}

func ColorizeActionSymbol(action string) string {
	switch action {
	case "create", "add", "added":
		return colorGreen.Sprint("+")
	case "update", "change", "changed":
		return colorYellow.Sprint("~")
	case "replace":
		return colorYellow.Sprint("±")
	case "delete", "destroy", "destroyed":
		return colorRed.Sprint("-")
	case "read":
		return colorCyan.Sprint("←")
	case "no-op":
		return colorGray.Sprint(" ")
	default:
		return " "
	}
}

func FormatDiff(before, after string) string {
	if before == "" && after == "" {
		return ""
	}

	var lines []string

	if before != "" {
		beforeLines := strings.Split(before, "\n")
		for _, line := range beforeLines {
			lines = append(lines, colorRed.Sprintf("- %s", line))
		}
	}

	if after != "" {
		afterLines := strings.Split(after, "\n")
		for _, line := range afterLines {
			lines = append(lines, colorGreen.Sprintf("+ %s", line))
		}
	}

	return strings.Join(lines, "\n")
}

func FormatUnifiedDiff(before, after string) string {
	if before == "" && after == "" {
		return ""
	}

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	var result []string

	beforeSet := make(map[string]bool)
	for _, line := range beforeLines {
		if line != "" {
			beforeSet[line] = true
		}
	}

	afterSet := make(map[string]bool)
	for _, line := range afterLines {
		if line != "" {
			afterSet[line] = true
		}
	}

	for _, line := range beforeLines {
		if line == "" {
			continue
		}
		if !afterSet[line] {
			result = append(result, colorRed.Sprintf("- %s", line))
		}
	}

	for _, line := range afterLines {
		if line == "" {
			continue
		}
		if !beforeSet[line] {
			result = append(result, colorGreen.Sprintf("+ %s", line))
		} else {
			result = append(result, fmt.Sprintf("  %s", line))
		}
	}

	return strings.Join(result, "\n")
}

func IndentString(s string, indent int) string {
	if s == "" {
		return s
	}

	prefix := strings.Repeat(" ", indent)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

func FormatResourceHeader(resourceType, name, action string) string {
	symbol := ColorizeActionSymbol(action)
	coloredAction := ColorizeAction(action)
	return fmt.Sprintf("%s %s %s (%s)", symbol, resourceType, colorBold.Sprint(name), coloredAction)
}

func FormatSectionHeader(title string) string {
	return colorBold.Sprintf("\n%s\n%s", title, strings.Repeat("─", len(title)))
}

func FormatKeyValue(key string, value any, indent int) string {
	prefix := strings.Repeat(" ", indent)
	return fmt.Sprintf("%s%s: %v", prefix, key, value)
}

func FormatChangeBlock(before, after any) string {
	var lines []string

	if before != nil {
		beforeStr := formatValue(before)
		if beforeStr != "" {
			lines = append(lines, colorRed.Sprintf("  - before: %s", beforeStr))
		}
	}

	if after != nil {
		afterStr := formatValue(after)
		if afterStr != "" {
			lines = append(lines, colorGreen.Sprintf("  + after:  %s", afterStr))
		}
	}

	return strings.Join(lines, "\n")
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		if len(val) > 100 {
			return TruncateString(val, 100)
		}
		return val
	case map[string]any:
		return fmt.Sprintf("<%d fields>", len(val))
	case []any:
		return fmt.Sprintf("[%d items]", len(val))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func IsDestructiveAction(action string) bool {
	switch action {
	case "delete", "destroy", "destroyed", "replace":
		return true
	default:
		return false
	}
}

func IsCreativeAction(action string) bool {
	switch action {
	case "create", "add", "added":
		return true
	default:
		return false
	}
}

func IsModifyAction(action string) bool {
	switch action {
	case "update", "change", "changed", "replace":
		return true
	default:
		return false
	}
}
