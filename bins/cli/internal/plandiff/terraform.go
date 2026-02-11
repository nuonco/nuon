package plandiff

import (
	"fmt"
	"strings"
)

// ParseTerraformPlan parses a TerraformPlan into a structured ParsedTerraformPlan
func ParseTerraformPlan(plan *TerraformPlan) *ParsedTerraformPlan {
	parsed := &ParsedTerraformPlan{}

	// Parse resource drift
	if plan.ResourceDrift != nil {
		for _, rd := range plan.ResourceDrift {
			mergedAfter := mergeAfterUnknown(rd.Change.After, rd.Change.AfterUnknown)

			for _, action := range rd.Change.Actions {
				incrementSummary(&parsed.Drift.Summary, action)
				if action == TerraformActionReplace {
					incrementSummary(&parsed.Drift.Summary, TerraformActionDelete)
					incrementSummary(&parsed.Drift.Summary, TerraformActionCreate)
				}

				parsed.Drift.Changes = append(parsed.Drift.Changes, ParsedTerraformResourceChange{
					Address:  rd.Address,
					Module:   rd.ModuleAddress,
					Resource: rd.Type,
					Name:     rd.Name,
					Action:   action,
					Before:   rd.Change.Before,
					After:    mergedAfter,
				})
			}
		}
	}

	// Parse resource changes
	for _, rc := range plan.ResourceChanges {
		mergedAfter := mergeAfterUnknown(rc.Change.After, rc.Change.AfterUnknown)

		// Handle read-only data sources
		if len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == TerraformActionRead {
			incrementSummary(&parsed.Resources.Summary, TerraformActionRead)
			parsed.Resources.Changes = append(parsed.Resources.Changes, ParsedTerraformResourceChange{
				Address:  rc.Address,
				Module:   rc.ModuleAddress,
				Resource: rc.Type,
				Name:     rc.Name,
				Action:   TerraformActionRead,
				Before:   rc.Change.Before,
				After:    mergedAfter,
			})
			continue
		}

		for _, action := range rc.Change.Actions {
			incrementSummary(&parsed.Resources.Summary, action)
			if action == TerraformActionReplace {
				incrementSummary(&parsed.Resources.Summary, TerraformActionDelete)
				incrementSummary(&parsed.Resources.Summary, TerraformActionCreate)
			}

			parsed.Resources.Changes = append(parsed.Resources.Changes, ParsedTerraformResourceChange{
				Address:  rc.Address,
				Module:   rc.ModuleAddress,
				Resource: rc.Type,
				Name:     rc.Name,
				Action:   action,
				Before:   rc.Change.Before,
				After:    mergedAfter,
			})
		}
	}

	// Parse output changes
	if plan.OutputChanges != nil {
		for output, oc := range plan.OutputChanges {
			mergedAfter := mergeAfterUnknown(oc.After, oc.AfterUnknown)

			// Handle read-only outputs
			if len(oc.Actions) == 1 && oc.Actions[0] == TerraformActionRead {
				incrementSummary(&parsed.Outputs.Summary, TerraformActionRead)
				parsed.Outputs.Changes = append(parsed.Outputs.Changes, TerraformOutputChange{
					Output:          output,
					Action:          TerraformActionRead,
					Before:          oc.Before,
					After:           mergedAfter,
					AfterUnknown:    oc.AfterUnknown,
					AfterSensitive:  oc.AfterSensitive,
					BeforeSensitive: oc.BeforeSensitive,
				})
				continue
			}

			for _, action := range oc.Actions {
				incrementSummary(&parsed.Outputs.Summary, action)
				if action == TerraformActionReplace {
					incrementSummary(&parsed.Outputs.Summary, TerraformActionDelete)
					incrementSummary(&parsed.Outputs.Summary, TerraformActionCreate)
				}

				parsed.Outputs.Changes = append(parsed.Outputs.Changes, TerraformOutputChange{
					Output:          output,
					Action:          action,
					Before:          oc.Before,
					After:           mergedAfter,
					AfterUnknown:    oc.AfterUnknown,
					AfterSensitive:  oc.AfterSensitive,
					BeforeSensitive: oc.BeforeSensitive,
				})
			}
		}
	}

	return parsed
}

// incrementSummary increments the appropriate counter in the summary based on action
func incrementSummary(summary *Summary, action TerraformChangeAction) {
	switch action {
	case TerraformActionCreate:
		summary.Create++
	case TerraformActionUpdate:
		summary.Update++
	case TerraformActionDelete:
		summary.Delete++
	case TerraformActionReplace:
		summary.Replace++
	case TerraformActionRead:
		summary.Read++
	case TerraformActionNoOp:
		summary.NoOp++
	}
}

// mergeAfterUnknown merges after values with after_unknown markers
// Unknown values are replaced with "Known after apply" placeholder
func mergeAfterUnknown(after, afterUnknown any) any {
	if afterUnknown == nil {
		return after
	}

	unknownMap, ok := afterUnknown.(map[string]any)
	if !ok {
		return after
	}

	var merged map[string]any
	if after != nil {
		if afterMap, ok := after.(map[string]any); ok {
			merged = make(map[string]any)
			for k, v := range afterMap {
				merged[k] = v
			}
		} else {
			return after
		}
	} else {
		merged = make(map[string]any)
	}

	processUnknown(merged, unknownMap)
	return merged
}

// processUnknown recursively processes unknown markers
func processUnknown(target, unknown map[string]any) {
	for key, value := range unknown {
		switch v := value.(type) {
		case bool:
			if v {
				target[key] = "(known after apply)"
			}
		case map[string]any:
			if targetChild, ok := target[key].(map[string]any); ok {
				processUnknown(targetChild, v)
			} else {
				child := make(map[string]any)
				processUnknown(child, v)
				target[key] = child
			}
		}
	}
}

// FormatTerraformPlan formats a parsed Terraform plan for terminal output
func FormatTerraformPlan(parsed *ParsedTerraformPlan) string {
	var sb strings.Builder

	// Calculate total summary for the header
	totalSummary := Summary{
		Create:  parsed.Resources.Summary.Create,
		Update:  parsed.Resources.Summary.Update,
		Delete:  parsed.Resources.Summary.Delete,
		Replace: parsed.Resources.Summary.Replace,
		Read:    parsed.Resources.Summary.Read,
		NoOp:    parsed.Resources.Summary.NoOp,
	}

	sb.WriteString(FormatSummary(totalSummary))
	sb.WriteString("\n")

	// Format drift section if there are any drift changes
	if len(parsed.Drift.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Resource Drift"))
		sb.WriteString("\n")
		sb.WriteString(formatResourceChanges(parsed.Drift.Changes))
		sb.WriteString("\n")
	}

	// Format resource changes
	if len(parsed.Resources.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Resource Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatResourceChanges(parsed.Resources.Changes))
		sb.WriteString("\n")
	}

	// Format output changes
	if len(parsed.Outputs.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Output Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatOutputChanges(parsed.Outputs.Changes))
	}

	return sb.String()
}

// formatResourceChanges formats a list of resource changes
func formatResourceChanges(changes []ParsedTerraformResourceChange) string {
	var sb strings.Builder

	for _, change := range changes {
		// Skip no-op changes in output
		if change.Action == TerraformActionNoOp {
			continue
		}

		sb.WriteString(FormatResourceHeader(change.Resource, change.Address, string(change.Action)))
		sb.WriteString("\n")

		// Show module if present
		if change.Module != nil && *change.Module != "" {
			sb.WriteString(fmt.Sprintf("    module: %s\n", *change.Module))
		}

		// Show before/after for changes using Terraform-style field diff
		if change.Action != TerraformActionRead {
			diffOutput := formatTerraformFieldDiff(change.Before, change.After, change.Action)
			if diffOutput != "" {
				sb.WriteString(diffOutput)
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// formatTerraformFieldDiff formats before/after values in Terraform CLI style
// showing each field with +, -, or ~ prefixes
func formatTerraformFieldDiff(before, after any, action TerraformChangeAction) string {
	var lines []string

	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)

	// If both are maps, do a field-by-field diff
	if beforeIsMap || afterIsMap {
		if beforeMap == nil {
			beforeMap = make(map[string]any)
		}
		if afterMap == nil {
			afterMap = make(map[string]any)
		}

		// Collect all keys from both maps
		allKeys := make(map[string]bool)
		for k := range beforeMap {
			allKeys[k] = true
		}
		for k := range afterMap {
			allKeys[k] = true
		}

		// Sort keys for consistent output
		sortedKeys := make([]string, 0, len(allKeys))
		for k := range allKeys {
			sortedKeys = append(sortedKeys, k)
		}
		sortStrings(sortedKeys)

		for _, key := range sortedKeys {
			beforeVal, hasBefore := beforeMap[key]
			afterVal, hasAfter := afterMap[key]

			line := formatFieldChange(key, beforeVal, afterVal, hasBefore, hasAfter, 4)
			if line != "" {
				lines = append(lines, line)
			}
		}
	} else {
		// Simple value comparison
		if before != nil && after != nil {
			if fmt.Sprintf("%v", before) != fmt.Sprintf("%v", after) {
				lines = append(lines, colorRed.Sprintf("    - %v", formatSimpleValue(before, 4)))
				lines = append(lines, colorGreen.Sprintf("    + %v", formatSimpleValue(after, 4)))
			}
		} else if before != nil {
			lines = append(lines, colorRed.Sprintf("    - %v", formatSimpleValue(before, 4)))
		} else if after != nil {
			lines = append(lines, colorGreen.Sprintf("    + %v", formatSimpleValue(after, 4)))
		}
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// formatFieldChange formats a single field change with appropriate prefix
func formatFieldChange(key string, before, after any, hasBefore, hasAfter bool, indent int) string {
	prefix := strings.Repeat(" ", indent)

	// Field was removed
	if hasBefore && !hasAfter {
		return colorRed.Sprintf("%s- %s = %s", prefix, key, formatSimpleValue(before, indent+2))
	}

	// Field was added
	if !hasBefore && hasAfter {
		return colorGreen.Sprintf("%s+ %s = %s", prefix, key, formatSimpleValue(after, indent+2))
	}

	// Field exists in both - check if changed
	if hasBefore && hasAfter {
		beforeStr := formatSimpleValue(before, indent+2)
		afterStr := formatSimpleValue(after, indent+2)

		if beforeStr != afterStr {
			return colorYellow.Sprintf("%s~ %s = %s -> %s", prefix, key, beforeStr, afterStr)
		}
		// Unchanged - don't output
		return ""
	}

	return ""
}

// formatSimpleValue formats a value for display, handling nested structures
// indent is used for recursive formatting of nested maps and arrays
func formatSimpleValue(v any, indent int) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		// Check if it's an integer
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		return formatMapValue(val, indent)
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		return formatArrayValue(val, indent)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatMapValue formats a map with proper indentation
func formatMapValue(m map[string]any, indent int) string {
	var lines []string
	prefix := strings.Repeat(" ", indent)

	// Collect and sort keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)

	lines = append(lines, "{")
	for _, k := range keys {
		v := m[k]
		valStr := formatSimpleValue(v, indent+2)
		lines = append(lines, fmt.Sprintf("%s  %s = %s", prefix, k, valStr))
	}
	lines = append(lines, prefix+"}")

	return strings.Join(lines, "\n")
}

// formatArrayValue formats an array with proper indentation
func formatArrayValue(arr []any, indent int) string {
	var lines []string
	prefix := strings.Repeat(" ", indent)

	lines = append(lines, "[")
	for _, item := range arr {
		valStr := formatSimpleValue(item, indent+2)
		lines = append(lines, fmt.Sprintf("%s  %s,", prefix, valStr))
	}
	lines = append(lines, prefix+"]")

	return strings.Join(lines, "\n")
}

// sortStrings sorts a slice of strings in place
func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// formatOutputChanges formats a list of output changes
func formatOutputChanges(changes []TerraformOutputChange) string {
	var sb strings.Builder

	for _, change := range changes {
		// Skip no-op changes in output
		if change.Action == TerraformActionNoOp {
			continue
		}

		sb.WriteString(FormatResourceHeader("output", change.Output, string(change.Action)))
		sb.WriteString("\n")

		// Handle sensitive values
		if isSensitive(change.BeforeSensitive) || isSensitive(change.AfterSensitive) {
			sb.WriteString("    (sensitive value)\n")
		} else if change.Action != TerraformActionRead {
			diffOutput := formatTerraformFieldDiff(change.Before, change.After, change.Action)
			if diffOutput != "" {
				sb.WriteString(diffOutput)
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// isSensitive checks if a value indicates sensitive data
func isSensitive(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// HasChanges returns true if the parsed plan has any non-no-op changes
func HasTerraformChanges(parsed *ParsedTerraformPlan) bool {
	return parsed.Resources.Summary.Create > 0 ||
		parsed.Resources.Summary.Update > 0 ||
		parsed.Resources.Summary.Delete > 0 ||
		parsed.Resources.Summary.Replace > 0 ||
		parsed.Outputs.Summary.Create > 0 ||
		parsed.Outputs.Summary.Update > 0 ||
		parsed.Outputs.Summary.Delete > 0 ||
		parsed.Outputs.Summary.Replace > 0 ||
		parsed.Drift.Summary.Create > 0 ||
		parsed.Drift.Summary.Update > 0 ||
		parsed.Drift.Summary.Delete > 0 ||
		parsed.Drift.Summary.Replace > 0
}
