package workflow

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

type policyViolation struct {
	PolicyID   string `json:"policy_id"`
	PolicyName string `json:"policy_name"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
}

type policyNameLookup interface {
	PolicyNameByID(id string) string
}

func parsePolicyViolations(raw any) []policyViolation {
	var violations []policyViolation
	if raw == nil {
		return violations
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return violations
	}

	_ = json.Unmarshal(data, &violations)
	return violations
}

func extractPolicyViolations(metadata map[string]any) ([]policyViolation, []policyViolation) {
	var denyViolations []policyViolation
	var warnViolations []policyViolation
	if metadata == nil {
		return denyViolations, warnViolations
	}

	if denyRaw, ok := metadata["deny_violations"]; ok {
		denyViolations = parsePolicyViolations(denyRaw)
	}
	if warnRaw, ok := metadata["warn_violations"]; ok {
		warnViolations = parsePolicyViolations(warnRaw)
	}

	return denyViolations, warnViolations
}

func (m model) stepDetailViewPolicyViolations() string {
	if m.selectedStep == nil || m.selectedStep.Status == nil {
		return ""
	}

	denyViolations, warnViolations := extractPolicyViolations(m.selectedStep.Status.Metadata)
	if len(denyViolations) == 0 && len(warnViolations) == 0 {
		return ""
	}

	header := styles.TextBold.Render("Policy Violations")
	lookup := policyNameLookup(m)
	sections := []string{header}

	contentWidth := m.stepDetail.Width - 6

	if len(denyViolations) > 0 {
		headerText := styles.TextError.Render(fmt.Sprintf("Deny violations (%d)", len(denyViolations)))
		lines := []string{headerText, ""}
		for i, violation := range denyViolations {
			lines = append(lines, formatPolicyViolationLine(violation, lookup, contentWidth))
			if i < len(denyViolations)-1 {
				lines = append(lines, "")
			}
		}
		sections = append(sections, policyDenyStyle.Width(m.stepDetail.Width-2).Padding(1).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
	}

	if len(warnViolations) > 0 {
		headerText := styles.TextWarning.Render(fmt.Sprintf("Warnings (%d)", len(warnViolations)))
		lines := []string{headerText, ""}
		for i, violation := range warnViolations {
			lines = append(lines, formatPolicyViolationLine(violation, lookup, contentWidth))
			if i < len(warnViolations)-1 {
				lines = append(lines, "")
			}
		}
		sections = append(sections, policyWarnStyle.Width(m.stepDetail.Width-2).Padding(1).Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return policySectionStyle.Width(m.stepDetail.Width).Padding(1).Margin(0, 0, 1).Render(content)
}

func formatPolicyViolationLine(violation policyViolation, lookup policyNameLookup, width int) string {
	message := violation.Message
	if message == "" {
		if violation.Severity == "warn" {
			message = "Policy warning"
		} else {
			message = "Policy check failed"
		}
	}

	wrapWidth := uint(width - 4)
	if wrapWidth < 40 {
		wrapWidth = 40
	}
	wrapped := wordwrap.WrapString(message, wrapWidth)

	policyName := violation.PolicyName
	if policyName == "" {
		policyName = lookup.PolicyNameByID(violation.PolicyID)
	}
	if policyName != "" {
		label := policyName
		if violation.PolicyID != "" {
			label = fmt.Sprintf("%s (%s)", policyName, violation.PolicyID)
		}
		return lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("- %s", wrapped),
			styles.TextSubtle.Render(fmt.Sprintf("  [%s]", label)),
		)
	}
	return fmt.Sprintf("- %s", wrapped)
}
