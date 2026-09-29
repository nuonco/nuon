package plandiff

import (
	"fmt"
	"strings"
)

func ParseHelmPlan(plan *HelmPlan) *ParsedHelmPlan {
	parsed := &ParsedHelmPlan{}

	for _, item := range plan.HelmContentDiff {
		resource := item.Kind
		if item.API != "" {
			resource = fmt.Sprintf("%s.%s", item.API, item.Kind)
		}

		var before, after *string
		var entryType int
		if len(item.Entries) > 0 {
			beforeParts := []string{}
			afterParts := []string{}
			for _, entry := range item.Entries {
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
			if item.Before != "" {
				before = &item.Before
			}
			if item.After != "" {
				after = &item.After
			}
		}

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

func determineHelmAction(entryType int, before, after *string) HelmK8sChangeAction {
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

	return HelmK8sActionChanged
}

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

func FormatHelmPlan(parsed *ParsedHelmPlan) string {
	var sb strings.Builder

	sb.WriteString(FormatSummary(parsed.Summary))
	sb.WriteString("\n")

	if len(parsed.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Helm Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatHelmChanges(parsed.Changes))
	}

	return sb.String()
}

func formatHelmChanges(changes []ParsedHelmChange) string {
	var sb strings.Builder

	for _, change := range changes {
		resourceName := change.Release
		if change.Workspace != "" {
			resourceName = fmt.Sprintf("%s/%s", change.Workspace, change.Release)
		}

		sb.WriteString(FormatResourceHeader(change.ResourceType, resourceName, string(change.Action)))
		sb.WriteString("\n")

		if change.Resource != change.ResourceType {
			sb.WriteString(fmt.Sprintf("    api: %s\n", change.Resource))
		}

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

func HasHelmChanges(parsed *ParsedHelmPlan) bool {
	return parsed.Summary.Add > 0 ||
		parsed.Summary.Change > 0 ||
		parsed.Summary.Destroy > 0
}
