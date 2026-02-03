package bubbles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

type ComponentResult struct {
	ID           string
	Name         string
	BuildID      string
	Status       string // "built", "failed", "timeout"
	Success      bool
	DenyCount    int
	WarnCount    int
	PassCount    int
	PolicyReport string // truncated report ID
	FailReason   string // reason for failure (if any)
}

type SyncResultsView struct {
	results []ComponentResult
}

func NewSyncResultsView() *SyncResultsView {
	return &SyncResultsView{
		results: make([]ComponentResult, 0),
	}
}

func (v *SyncResultsView) AddResult(result ComponentResult) {
	v.results = append(v.results, result)
}

func (v *SyncResultsView) Render() string {
	if len(v.results) == 0 {
		return ""
	}

	var lines []string

	maxNameLen := 0
	for _, r := range v.results {
		if len(r.Name) > maxNameLen {
			maxNameLen = len(r.Name)
		}
	}
	if maxNameLen < 10 {
		maxNameLen = 10
	}
	if maxNameLen > 30 {
		maxNameLen = 30
	}

	successIcon := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true).Render("✓")
	failIcon := lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true).Render("✗")
	headerStyle := lipgloss.NewStyle().Foreground(styles.SubtleColor).Bold(true)
	denyStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
	warnStyle := lipgloss.NewStyle().Foreground(styles.WarningColor)
	infoStyle := lipgloss.NewStyle().Foreground(styles.InfoColor)
	subtleStyle := lipgloss.NewStyle().Foreground(styles.SubtleColor)

	lines = append(lines, "")

	header := fmt.Sprintf("  %-*s  %-8s  %3s %3s %4s   %s",
		maxNameLen, "NAME", "STATUS", "ERR", "WRN", "PASS", "REPORT")
	lines = append(lines, headerStyle.Render(header))

	for _, r := range v.results {
		icon := successIcon
		if !r.Success {
			icon = failIcon
		}

		name := r.Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}
		name = fmt.Sprintf("%-*s", maxNameLen, name)
		status := fmt.Sprintf("%-8s", r.Status)

		denyText := denyStyle.Render(fmt.Sprintf("%3d", r.DenyCount))
		warnText := warnStyle.Render(fmt.Sprintf("%3d", r.WarnCount))
		passText := infoStyle.Render(fmt.Sprintf("%4d", r.PassCount))

		reportID := subtleStyle.Render("—")
		if r.PolicyReport != "" {
			reportID = subtleStyle.Render(r.PolicyReport)
		}

		line := fmt.Sprintf("%s %s  %s  %s %s %s   %s",
			icon, name, status, denyText, warnText, passText, reportID)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, renderSeparator())
	lines = append(lines, v.renderSummary())

	failedComponents := v.getFailedComponents()
	if len(failedComponents) > 0 {
		lines = append(lines, "")
		lines = append(lines, v.renderFailedSection(failedComponents))
	}

	if v.hasPolicyReports() {
		lines = append(lines, "")
		lines = append(lines, subtleStyle.Render("View details: nuon policies reports get -r <report-id>"))
	}

	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func (v *SyncResultsView) renderSummary() string {
	var builtCount, failedCount, totalDeny, totalWarn, totalPass int

	for _, r := range v.results {
		if r.Success {
			builtCount++
		} else {
			failedCount++
		}
		totalDeny += r.DenyCount
		totalWarn += r.WarnCount
		totalPass += r.PassCount
	}

	denyStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
	warnStyle := lipgloss.NewStyle().Foreground(styles.WarningColor)
	infoStyle := lipgloss.NewStyle().Foreground(styles.InfoColor)
	subtleStyle := lipgloss.NewStyle().Foreground(styles.SubtleColor)

	var parts []string
	if failedCount > 0 {
		successStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor)
		failStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
		parts = append(parts, successStyle.Render(fmt.Sprintf("%d built", builtCount)))
		parts = append(parts, failStyle.Render(fmt.Sprintf("%d failed", failedCount)))
	} else {
		parts = append(parts, fmt.Sprintf("%d components", len(v.results)))
	}

	parts = append(parts, denyStyle.Render(fmt.Sprintf("%d errors", totalDeny)))
	parts = append(parts, warnStyle.Render(fmt.Sprintf("%d warnings", totalWarn)))
	parts = append(parts, infoStyle.Render(fmt.Sprintf("%d info", totalPass)))

	return subtleStyle.Render(strings.Join(parts, " • "))
}

func (v *SyncResultsView) renderFailedSection(failed []ComponentResult) string {
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Failed components:"))

	for _, r := range failed {
		reason := r.FailReason
		if reason == "" {
			if r.DenyCount > 0 {
				reason = fmt.Sprintf("%d policy error(s) blocked build", r.DenyCount)
			} else {
				reason = "build failed"
			}
		}
		line := fmt.Sprintf("  %s    %s", r.Name, lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(reason))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (v *SyncResultsView) getFailedComponents() []ComponentResult {
	var failed []ComponentResult
	for _, r := range v.results {
		if !r.Success {
			failed = append(failed, r)
		}
	}
	return failed
}

func (v *SyncResultsView) hasPolicyReports() bool {
	for _, r := range v.results {
		if r.PolicyReport != "" {
			return true
		}
	}
	return false
}

func (v *SyncResultsView) HasErrors() bool {
	for _, r := range v.results {
		if !r.Success {
			return true
		}
	}
	return false
}

func renderSeparator() string {
	return lipgloss.NewStyle().Foreground(styles.SubtleColor).Render("────────────────────────────────────────────────────")
}

func truncateReportID(id string) string {
	if len(id) <= 14 {
		return id
	}
	return id[:12] + "..."
}
