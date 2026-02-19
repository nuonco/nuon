package plandiff

import (
	"fmt"
	"strings"
)

// ParseHelmPlan parses a HelmPlan into a structured ParsedHelmPlan
func ParseHelmPlan(plan *HelmPlan) *ParsedHelmPlan {
	parsed := &ParsedHelmPlan{}

	for _, item := range plan.HelmContentDiff {
		// Build resource identifier from Kind and Name
		resource := item.Kind
		if item.API != "" {
			resource = fmt.Sprintf("%s.%s", item.API, item.Kind)
		}

		// Build before/after from entries if available
		var before, after *string
		var entryType int
		if len(item.Entries) > 0 {
			beforeParts := []string{}
			afterParts := []string{}
			for _, entry := range item.Entries {
				// Use the first entry's type to determine action
				if entryType == 0 {
					entryType = entry.Type
				}
				if entry.Original != "" {
					beforeParts = append(beforeParts, fmt.Sprintf("%s: %s", entry.Path, entry.Original))
				}
				if entry.Applied != "" {
					afterParts = append(afterParts, fmt.Sprintf("%s: %s", entry.Path, entry.Applied))
				}
			}
			if len(beforeParts) > 0 {
				b := strings.Join(beforeParts, "\n")
				before = &b
			}
			if len(afterParts) > 0 {
				a := strings.Join(afterParts, "\n")
				after = &a
			}
		} else {
			// Use Before/After fields directly if no entries
			if item.Before != "" {
				before = &item.Before
			}
			if item.After != "" {
				after = &item.After
			}
		}

		// Determine action from entry type or infer from before/after
		action := determineHelmAction(entryType, before, after)
		incrementHelmSummary(&parsed.Summary, action)

		parsed.Changes = append(parsed.Changes, ParsedHelmChange{
			Workspace:    item.Namespace,
			Release:      item.Name,
			Resource:     resource,
			ResourceType: item.Kind,
			Action:       action,
			Before:       before,
			After:        after,
		})
	}

	return parsed
}

// determineHelmAction determines the action from entry type or infers from before/after
// Entry Type values: 1=add, 2=delete, 3=change
func determineHelmAction(entryType int, before, after *string) HelmK8sChangeAction {
	// If we have an entry type, use it
	if entryType > 0 {
		switch entryType {
		case 1:
			return HelmK8sActionAdded
		case 2:
			return HelmK8sActionDestroyed
		case 3:
			return HelmK8sActionChanged
		}
	}

	// Infer from before/after
	hasBefore := before != nil && *before != ""
	hasAfter := after != nil && *after != ""

	if !hasBefore && hasAfter {
		return HelmK8sActionAdded
	}
	if hasBefore && !hasAfter {
		return HelmK8sActionDestroyed
	}
	if hasBefore && hasAfter {
		return HelmK8sActionChanged
	}

	// Default to change if we can't determine
	return HelmK8sActionChanged
}

// incrementHelmSummary increments the appropriate counter in the summary based on action
func incrementHelmSummary(summary *Summary, action HelmK8sChangeAction) {
	switch action {
	case HelmK8sActionAdd, HelmK8sActionAdded:
		summary.Add++
	case HelmK8sActionChange, HelmK8sActionChanged:
		summary.Change++
	case HelmK8sActionDestroy, HelmK8sActionDestroyed:
		summary.Destroy++
	}
}

// FormatHelmPlan formats a parsed Helm plan for terminal output
func FormatHelmPlan(parsed *ParsedHelmPlan) string {
	var sb strings.Builder

	// Show summary
	sb.WriteString(FormatSummary(parsed.Summary))
	sb.WriteString("\n")

	// Format changes
	if len(parsed.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Helm Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatHelmChanges(parsed.Changes))
	}

	return sb.String()
}

// formatHelmChanges formats a list of Helm changes
func formatHelmChanges(changes []ParsedHelmChange) string {
	var sb strings.Builder

	for _, change := range changes {
		// Format resource header with namespace/name
		resourceName := change.Release
		if change.Workspace != "" {
			resourceName = fmt.Sprintf("%s/%s", change.Workspace, change.Release)
		}

		sb.WriteString(FormatResourceHeader(change.ResourceType, resourceName, string(change.Action)))
		sb.WriteString("\n")

		// Show resource details
		if change.Resource != change.ResourceType {
			sb.WriteString(fmt.Sprintf("    api: %s\n", change.Resource))
		}

		// Show before/after diff
		if change.Before != nil || change.After != nil {
			var before, after string
			if change.Before != nil {
				before = *change.Before
			}
			if change.After != nil {
				after = *change.After
			}

			diff := FormatDiff(before, after)
			if diff != "" {
				sb.WriteString(IndentString(diff, 4))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// HasHelmChanges returns true if the parsed plan has any changes
func HasHelmChanges(parsed *ParsedHelmPlan) bool {
	return parsed.Summary.Add > 0 ||
		parsed.Summary.Change > 0 ||
		parsed.Summary.Destroy > 0
}
