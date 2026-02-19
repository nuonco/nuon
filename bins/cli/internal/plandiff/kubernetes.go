package plandiff

import (
	"fmt"
	"strings"
)

// ParseKubernetesPlan parses a KubernetesPlan into a structured ParsedKubernetesPlan
func ParseKubernetesPlan(plan *KubernetesPlan) *ParsedKubernetesPlan {
	parsed := &ParsedKubernetesPlan{}

	for _, item := range plan.K8sContentDiff {
		// Handle errors separately
		if item.Error != "" {
			parsed.Errors = append(parsed.Errors, ParsedKubernetesError{
				Namespace:    item.Namespace,
				Name:         item.Name,
				Resource:     item.Resource,
				ResourceType: item.Kind,
				Error:        item.Error,
			})
			continue
		}

		// Build resource identifier from Kind and API
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
		}

		// Determine action from item Type field or infer from before/after
		action := determineKubernetesAction(item.Type, entryType, before, after)
		incrementKubernetesSummary(&parsed.Summary, action)

		parsed.Changes = append(parsed.Changes, ParsedKubernetesChange{
			Namespace:    item.Namespace,
			Name:         item.Name,
			Resource:     resource,
			ResourceType: item.Kind,
			Action:       action,
			Before:       before,
			After:        after,
		})
	}

	return parsed
}

// determineKubernetesAction determines the action from item type, entry type, or infers from before/after
// Item/Entry Type values: 1=add, 2=delete, 3=change
func determineKubernetesAction(itemType, entryType int, before, after *string) HelmK8sChangeAction {
	// Use item type first if available
	if itemType > 0 {
		switch itemType {
		case 1:
			return HelmK8sActionAdded
		case 2:
			return HelmK8sActionDestroyed
		case 3:
			return HelmK8sActionChanged
		}
	}

	// Fall back to entry type if item type is not set
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

// incrementKubernetesSummary increments the appropriate counter in the summary based on action
func incrementKubernetesSummary(summary *Summary, action HelmK8sChangeAction) {
	switch action {
	case HelmK8sActionAdd, HelmK8sActionAdded:
		summary.Add++
	case HelmK8sActionChange, HelmK8sActionChanged:
		summary.Change++
	case HelmK8sActionDestroy, HelmK8sActionDestroyed:
		summary.Destroy++
	}
}

// FormatKubernetesPlan formats a parsed Kubernetes plan for terminal output
func FormatKubernetesPlan(parsed *ParsedKubernetesPlan, planText string) string {
	var sb strings.Builder

	// Show plan text if present
	if planText != "" {
		sb.WriteString(colorBold.Sprint("Plan: "))
		sb.WriteString(planText)
		sb.WriteString("\n\n")
	}

	// Show summary
	sb.WriteString(FormatSummary(parsed.Summary))
	sb.WriteString("\n")

	// Format changes
	if len(parsed.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Kubernetes Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatKubernetesChanges(parsed.Changes))
	}

	// Format errors
	if len(parsed.Errors) > 0 {
		sb.WriteString(FormatSectionHeader("Errors"))
		sb.WriteString("\n")
		sb.WriteString(formatKubernetesErrors(parsed.Errors))
	}

	return sb.String()
}

// formatKubernetesChanges formats a list of Kubernetes changes
func formatKubernetesChanges(changes []ParsedKubernetesChange) string {
	var sb strings.Builder

	for _, change := range changes {
		// Format resource header with namespace/name
		resourceName := change.Name
		if change.Namespace != "" {
			resourceName = fmt.Sprintf("%s/%s", change.Namespace, change.Name)
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

// formatKubernetesErrors formats a list of Kubernetes errors
func formatKubernetesErrors(errors []ParsedKubernetesError) string {
	var sb strings.Builder

	for _, err := range errors {
		// Format resource header with namespace/name
		resourceName := err.Name
		if err.Namespace != "" {
			resourceName = fmt.Sprintf("%s/%s", err.Namespace, err.Name)
		}

		sb.WriteString(colorRed.Sprintf("✗ %s %s\n", err.ResourceType, colorBold.Sprint(resourceName)))

		// Show resource details
		if err.Resource != "" && err.Resource != err.ResourceType {
			sb.WriteString(fmt.Sprintf("    resource: %s\n", err.Resource))
		}

		// Show error message
		sb.WriteString(colorRed.Sprintf("    error: %s\n", err.Error))
		sb.WriteString("\n")
	}

	return sb.String()
}

// HasKubernetesChanges returns true if the parsed plan has any changes
func HasKubernetesChanges(parsed *ParsedKubernetesPlan) bool {
	return parsed.Summary.Add > 0 ||
		parsed.Summary.Change > 0 ||
		parsed.Summary.Destroy > 0 ||
		len(parsed.Errors) > 0
}
