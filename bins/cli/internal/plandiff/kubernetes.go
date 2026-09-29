package plandiff

import (
	"fmt"
	"strings"
)

func ParseKubernetesPlan(plan *KubernetesPlan) *ParsedKubernetesPlan {
	parsed := &ParsedKubernetesPlan{}

	for _, item := range plan.K8sContentDiff {
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
		}

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

func determineKubernetesAction(itemType, entryType int, before, after *string) HelmK8sChangeAction {
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

func FormatKubernetesPlan(parsed *ParsedKubernetesPlan, planText string) string {
	var sb strings.Builder

	if planText != "" {
		sb.WriteString(colorBold.Sprint("Plan: "))
		sb.WriteString(planText)
		sb.WriteString("\n\n")
	}

	sb.WriteString(FormatSummary(parsed.Summary))
	sb.WriteString("\n")

	if len(parsed.Changes) > 0 {
		sb.WriteString(FormatSectionHeader("Kubernetes Changes"))
		sb.WriteString("\n")
		sb.WriteString(formatKubernetesChanges(parsed.Changes))
	}

	if len(parsed.Errors) > 0 {
		sb.WriteString(FormatSectionHeader("Errors"))
		sb.WriteString("\n")
		sb.WriteString(formatKubernetesErrors(parsed.Errors))
	}

	return sb.String()
}

func formatKubernetesChanges(changes []ParsedKubernetesChange) string {
	var sb strings.Builder

	for _, change := range changes {
		resourceName := change.Name
		if change.Namespace != "" {
			resourceName = fmt.Sprintf("%s/%s", change.Namespace, change.Name)
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

func formatKubernetesErrors(errors []ParsedKubernetesError) string {
	var sb strings.Builder

	for _, err := range errors {
		resourceName := err.Name
		if err.Namespace != "" {
			resourceName = fmt.Sprintf("%s/%s", err.Namespace, err.Name)
		}

		sb.WriteString(colorRed.Sprintf("✗ %s %s\n", err.ResourceType, colorBold.Sprint(resourceName)))

		if err.Resource != "" && err.Resource != err.ResourceType {
			sb.WriteString(fmt.Sprintf("    resource: %s\n", err.Resource))
		}

		sb.WriteString(colorRed.Sprintf("    error: %s\n", err.Error))
		sb.WriteString("\n")
	}

	return sb.String()
}

func HasKubernetesChanges(parsed *ParsedKubernetesPlan) bool {
	return parsed.Summary.Add > 0 ||
		parsed.Summary.Change > 0 ||
		parsed.Summary.Destroy > 0 ||
		len(parsed.Errors) > 0
}
