package plandiff

import (
	"fmt"
	"strings"
)

func ParseTerraformPlan(plan *TerraformPlan) *ParsedTerraformPlan {
	parsed := &ParsedTerraformPlan{}

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

	for _, rc := range plan.ResourceChanges {
		mergedAfter := mergeAfterUnknown(rc.Change.After, rc.Change.AfterUnknown)

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

	if plan.OutputChanges != nil {
		for output, oc := range plan.OutputChanges {
			mergedAfter := mergeAfterUnknown(oc.After, oc.AfterUnknown)

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

func FormatTerraformPlan(parsed *ParsedTerraformPlan) string {
	var sb strings.Builder

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

	if len(parsed.Drift.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Resource Drift"))
		sb.WriteString("\n")
		sb.WriteString(formatResourceChanges(parsed.Drift.Changes))
		sb.WriteString("\n")
	}

	if len(parsed.Resources.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Resource Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatResourceChanges(parsed.Resources.Changes))
		sb.WriteString("\n")
	}

	if len(parsed.Outputs.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Output Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatOutputChanges(parsed.Outputs.Changes))
	}

	return sb.String()
}

func formatResourceChanges(changes []ParsedTerraformResourceChange) string {
	var sb strings.Builder

	for _, change := range changes {
		if change.Action == TerraformActionNoOp {
			continue
		}

		sb.WriteString(FormatResourceHeader(change.Resource, change.Address, string(change.Action)))
		sb.WriteString("\n")

		if change.Module != nil && *change.Module != "" {
			sb.WriteString(fmt.Sprintf("    module: %s\n", *change.Module))
		}

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

func formatTerraformFieldDiff(before, after any, action TerraformChangeAction) string {
	var lines []string

	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)

	if beforeIsMap || afterIsMap {
		if beforeMap == nil {
			beforeMap = make(map[string]any)
		}
		if afterMap == nil {
			afterMap = make(map[string]any)
		}

		allKeys := make(map[string]bool)
		for k := range beforeMap {
			allKeys[k] = true
		}
		for k := range afterMap {
			allKeys[k] = true
		}

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

func formatFieldChange(key string, before, after any, hasBefore, hasAfter bool, indent int) string {
	prefix := strings.Repeat(" ", indent)

	if hasBefore && !hasAfter {
		return colorRed.Sprintf("%s- %s = %s", prefix, key, formatSimpleValue(before, indent+2))
	}

	if !hasBefore && hasAfter {
		return colorGreen.Sprintf("%s+ %s = %s", prefix, key, formatSimpleValue(after, indent+2))
	}

	if hasBefore && hasAfter {
		beforeStr := formatSimpleValue(before, indent+2)
		afterStr := formatSimpleValue(after, indent+2)

		if beforeStr != afterStr {
			return colorYellow.Sprintf("%s~ %s = %s -> %s", prefix, key, beforeStr, afterStr)
		}
		return ""
	}

	return ""
}

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

func formatMapValue(m map[string]any, indent int) string {
	var lines []string
	prefix := strings.Repeat(" ", indent)

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

func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func formatOutputChanges(changes []TerraformOutputChange) string {
	var sb strings.Builder

	for _, change := range changes {
		if change.Action == TerraformActionNoOp {
			continue
		}

		sb.WriteString(FormatResourceHeader("output", change.Output, string(change.Action)))
		sb.WriteString("\n")

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

func isSensitive(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

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
