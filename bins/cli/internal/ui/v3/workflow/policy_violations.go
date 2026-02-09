package workflow

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

type policyViolation struct {
	PolicyID string `json:"policy_id"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
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
	sections := []string{header}

	if len(denyViolations) > 0 {
		sections = append(sections, styles.TextError.Render(fmt.Sprintf("Deny violations (%d)", len(denyViolations))))
		for _, violation := range denyViolations {
			line := fmt.Sprintf("- %s", violation.Message)
			if violation.Message == "" {
				line = "- Policy check failed"
			}
			if violation.PolicyID != "" {
				line = fmt.Sprintf("%s %s", line, styles.TextSubtle.Render(fmt.Sprintf("(%s)", violation.PolicyID)))
			}
			sections = append(sections, line)
		}
	}

	if len(warnViolations) > 0 {
		sections = append(sections, styles.TextWarning.Render(fmt.Sprintf("Warnings (%d)", len(warnViolations))))
		for _, violation := range warnViolations {
			line := fmt.Sprintf("- %s", violation.Message)
			if violation.Message == "" {
				line = "- Policy warning"
			}
			if violation.PolicyID != "" {
				line = fmt.Sprintf("%s %s", line, styles.TextSubtle.Render(fmt.Sprintf("(%s)", violation.PolicyID)))
			}
			sections = append(sections, line)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return policySectionStyle.Width(m.stepDetail.Width).Padding(1).Margin(0, 0, 1).Render(content)
}
