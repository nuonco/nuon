package workflow

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (m *model) stepDetailViewBranchStep(step *models.AppWorkflowStep) string {
	if step.Status == nil || step.Status.Metadata == nil {
		return ""
	}
	meta := step.Status.Metadata

	switch {
	case step.Name == "fetch commit":
		return m.stepDetailViewFetchCommit(meta)
	case step.Name == "fetch app config":
		return m.stepDetailViewAppConfig(meta)
	case step.Name == "building components and sandbox":
		return m.stepDetailViewBuilds(meta)
	case strings.HasPrefix(step.Name, "deploy install group:"):
		return m.stepDetailViewInstallGroup(meta)
	default:
		return ""
	}
}

func (m model) stepDetailViewFetchCommit(meta map[string]interface{}) string {
	w := m.stepDetail.Width() - 4

	commitSHA := metaStr(meta, "commit_sha")
	commitMsg := metaStr(meta, "commit_message")
	authorName := metaStr(meta, "author_name")
	repo := metaStr(meta, "repo")
	branch := metaStr(meta, "branch")

	rows := []string{}
	if commitSHA != "" {
		display := commitSHA
		if len(display) > 12 {
			display = display[:12]
		}
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("commit"), styles.TextBold.Render(display)))
	}
	if commitMsg != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("message"), commitMsg))
	}
	if authorName != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("author"), authorName))
	}
	if repo != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("repo"), repo))
	}
	if branch != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("branch"), branch))
	}

	prURL := metaStr(meta, "pr_url")
	prTitle := metaStr(meta, "pr_title")
	prStatus := metaStr(meta, "pr_status")
	if prURL != "" || prTitle != "" {
		rows = append(rows, "")
		if prTitle != "" {
			rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("PR"), prTitle))
		}
		if prStatus != "" {
			rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("status"), prStatus))
		}
		if prURL != "" {
			rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("url"), styles.Link.Render(prURL)))
		}
	}

	filesChanged := metaStr(meta, "files_changed")
	additions := metaStr(meta, "additions")
	deletions := metaStr(meta, "deletions")
	if filesChanged != "" || additions != "" || deletions != "" {
		rows = append(rows, "")
		parts := []string{}
		if filesChanged != "" {
			parts = append(parts, fmt.Sprintf("%s files changed", filesChanged))
		}
		if additions != "" {
			parts = append(parts, lipgloss.NewStyle().Foreground(styles.SuccessColor).Render(fmt.Sprintf("+%s", additions)))
		}
		if deletions != "" {
			parts = append(parts, lipgloss.NewStyle().Foreground(styles.ErrorColor).Render(fmt.Sprintf("-%s", deletions)))
		}
		rows = append(rows, strings.Join(parts, "  "))
	}

	if len(rows) == 0 {
		return ""
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(1).Margin(0, 1).
		Width(w).
		Render(content)
}

func (m *model) stepDetailViewAppConfig(meta map[string]interface{}) string {
	w := m.stepDetail.Width() - 4
	addStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor)
	removeStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
	changeStyle := lipgloss.NewStyle().Foreground(styles.WarningColor)

	componentCount := metaStr(meta, "component_count")
	actionCount := metaStr(meta, "action_count")

	rows := []string{}
	if componentCount != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("components"), componentCount))
	}
	if actionCount != "" {
		rows = append(rows, fmt.Sprintf("%s  %s", styles.TextSubtle.Render("actions"), actionCount))
	}

	diffChanged := metaStr(meta, "diff_changed")
	diffAdditions := metaStr(meta, "diff_additions")
	diffRemovals := metaStr(meta, "diff_removals")

	if diffChanged != "" || diffAdditions != "" || diffRemovals != "" {
		rows = append(rows, "")
		parts := []string{}
		if diffAdditions != "" && diffAdditions != "0" {
			parts = append(parts, addStyle.Render(fmt.Sprintf("+%s added", diffAdditions)))
		}
		if diffRemovals != "" && diffRemovals != "0" {
			parts = append(parts, removeStyle.Render(fmt.Sprintf("-%s removed", diffRemovals)))
		}
		if diffChanged != "" && diffChanged != "0" {
			parts = append(parts, changeStyle.Render(fmt.Sprintf("~%s changed", diffChanged)))
		}
		if len(parts) > 0 {
			rows = append(rows, strings.Join(parts, "  "))
		} else {
			rows = append(rows, styles.TextSubtle.Render("no changes"))
		}
	}

	headerContent := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sections := []string{
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.SubtleColor).
			Padding(1).Margin(0, 1).
			Width(w).
			Render(headerContent),
	}

	rawSections, ok := meta["diff_sections"]
	if !ok {
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	sectionList := toSliceOfMaps(rawSections)
	if len(sectionList) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	m.diffSectionCount = len(sectionList)
	if m.diffSectionCursor >= m.diffSectionCount {
		m.diffSectionCursor = 0
	}

	if len(sectionList) > 0 {
		sections = append(sections, styles.TextSubtle.Padding(0, 2).Render("↑/↓ navigate  enter expand  esc collapse"))
	}

	expandedSections := m.expandedDiffSections
	for i, sec := range sectionList {
		name := metaMapStr(sec, "name")
		if name == "" {
			name = fmt.Sprintf("Section %d", i+1)
		}

		additions := metaMapInt(sec, "additions")
		removals := metaMapInt(sec, "removals")
		changed := metaMapInt(sec, "changed")

		countParts := []string{}
		if additions > 0 {
			countParts = append(countParts, addStyle.Render(fmt.Sprintf("+%d", additions)))
		}
		if removals > 0 {
			countParts = append(countParts, removeStyle.Render(fmt.Sprintf("-%d", removals)))
		}
		if changed > 0 {
			countParts = append(countParts, changeStyle.Render(fmt.Sprintf("~%d", changed)))
		}
		countStr := ""
		if len(countParts) > 0 {
			countStr = " (" + strings.Join(countParts, "/") + ")"
		}

		expanded := expandedSections[i]
		caret := "▸"
		if expanded {
			caret = "▾"
		}

		isCursor := i == m.diffSectionCursor
		cursor := "  "
		if isCursor {
			cursor = "▶ "
		}

		header := fmt.Sprintf("%s%s %s%s",
			cursor,
			caret,
			styles.TextBold.Render(name),
			countStr,
		)
		sectionRows := []string{header}

		if expanded {
			sectionRows = append(sectionRows, "")
			entryRows := renderDiffEntries(sec, addStyle, removeStyle, changeStyle)
			if len(entryRows) == 0 {
				sectionRows = append(sectionRows, styles.TextSubtle.Render("  no details available"))
			} else {
				sectionRows = append(sectionRows, entryRows...)
			}
		}

		borderColor := styles.SubtleColor
		if isCursor {
			borderColor = styles.AccentColor
		}

		sectionContent := lipgloss.JoinVertical(lipgloss.Left, sectionRows...)
		sections = append(sections, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1).Margin(0, 1).
			Width(w).
			Render(sectionContent))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

type diffParsedEntry struct {
	op   string
	name string
	desc string
}

func renderDiffEntries(sec map[string]interface{}, addStyle, removeStyle, changeStyle lipgloss.Style) []string {
	var parsed []diffParsedEntry

	for _, entry := range toSliceOfMaps(sec["entries"]) {
		parsed = append(parsed, diffParsedEntry{
			op:   metaMapStr(entry, "op"),
			name: metaMapStr(entry, "name"),
			desc: metaMapStr(entry, "description"),
		})
	}

	if len(parsed) == 0 {
		additions := metaMapInt(sec, "additions")
		removals := metaMapInt(sec, "removals")
		changed := metaMapInt(sec, "changed")
		var rows []string
		if additions > 0 {
			rows = append(rows, addStyle.Render(fmt.Sprintf("  + %d added", additions)))
		}
		if removals > 0 {
			rows = append(rows, removeStyle.Render(fmt.Sprintf("  - %d removed", removals)))
		}
		if changed > 0 {
			rows = append(rows, changeStyle.Render(fmt.Sprintf("  ~ %d changed", changed)))
		}
		if len(rows) == 0 {
			rows = append(rows, styles.TextSubtle.Render("  no changes"))
		}
		return rows
	}

	grouped := make(map[string][]diffParsedEntry)
	var orderedNames []string
	for _, e := range parsed {
		if _, seen := grouped[e.name]; !seen {
			orderedNames = append(orderedNames, e.name)
		}
		grouped[e.name] = append(grouped[e.name], e)
	}

	var rows []string
	for _, name := range orderedNames {
		group := grouped[name]

		allAdd := true
		allRemove := true
		for _, e := range group {
			if e.op != "add" {
				allAdd = false
			}
			if e.op != "remove" {
				allRemove = false
			}
		}

		itemType := extractType(group)
		typeLabel := ""
		if itemType != "" {
			typeLabel = " " + styles.TextSubtle.Render(fmt.Sprintf("[%s]", itemType))
		}

		if allAdd {
			rows = append(rows, addStyle.Render(fmt.Sprintf("  + %s", name))+typeLabel+" "+styles.TextSubtle.Render("(new)"))
			for _, e := range group {
				if e.desc != "" && !isTypeField(e.desc) {
					rows = append(rows, addStyle.Render(fmt.Sprintf("      + %s", e.desc)))
				}
			}
		} else if allRemove {
			rows = append(rows, removeStyle.Render(fmt.Sprintf("  - %s", name))+typeLabel+" "+styles.TextSubtle.Render("(deleted)"))
		} else {
			rows = append(rows, changeStyle.Render(fmt.Sprintf("  ~ %s", name))+typeLabel)
			for _, e := range group {
				if e.desc == "" || isTypeField(e.desc) {
					continue
				}
				switch e.op {
				case "add":
					rows = append(rows, addStyle.Render(fmt.Sprintf("      + %s", e.desc)))
				case "remove":
					rows = append(rows, removeStyle.Render(fmt.Sprintf("      - %s", e.desc)))
				case "change":
					rows = append(rows, changeStyle.Render(fmt.Sprintf("      ~ %s", e.desc)))
				default:
					rows = append(rows, styles.TextSubtle.Render(fmt.Sprintf("        %s", e.desc)))
				}
			}
		}
	}

	return rows
}

func extractType(entries []diffParsedEntry) string {
	for _, e := range entries {
		if strings.Contains(e.desc, "-> '") {
			parts := strings.SplitN(e.desc, "-> '", 2)
			if len(parts) == 2 {
				val := strings.TrimSuffix(parts[1], "'")
				if isComponentType(val) {
					return val
				}
			}
		}
		if strings.HasPrefix(e.desc, "'") && strings.HasSuffix(e.desc, "' (unchanged)") {
			val := strings.TrimPrefix(e.desc, "'")
			val = strings.TrimSuffix(val, "' (unchanged)")
			if isComponentType(val) {
				return val
			}
		}
	}
	return ""
}

func isComponentType(s string) bool {
	switch s {
	case "helm_chart", "terraform_module", "docker_build", "container_image",
		"kubernetes_manifest", "job", "pulumi", "external_image":
		return true
	}
	return false
}

func isTypeField(desc string) bool {
	for _, t := range []string{"helm_chart", "terraform_module", "docker_build", "container_image",
		"kubernetes_manifest", "job", "pulumi", "external_image"} {
		if strings.Contains(desc, "'"+t+"'") {
			return true
		}
	}
	return false
}

func (m *model) stepDetailViewBuilds(meta map[string]interface{}) string {
	w := m.stepDetail.Width() - 4

	buildList := toSliceOfMaps(meta["builds"])
	if len(buildList) == 0 {
		return ""
	}

	m.diffSectionCount = len(buildList)
	if m.diffSectionCursor >= m.diffSectionCount {
		m.diffSectionCursor = 0
	}

	rows := []string{
		styles.TextBold.Render("Component Builds"),
		styles.TextSubtle.Render("↑/↓ navigate  enter expand  esc collapse"),
		"",
	}

	for i, build := range buildList {
		name := metaMapStr(build, "component_name")
		status := metaMapStr(build, "status")
		cacheStatus := metaMapStr(build, "cache_status")
		compType := metaMapStr(build, "component_type")
		isNew := metaMapStr(build, "is_new") == "true"
		if name == "" {
			name = metaMapStr(build, "component_id")
		}

		icon := "○"
		statusStyle := styles.TextSubtle
		switch status {
		case "success":
			icon = "✓"
			statusStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor)
		case "error":
			icon = "✗"
			statusStyle = lipgloss.NewStyle().Foreground(styles.ErrorColor)
		case "in-progress":
			icon = m.spinner.View()
			statusStyle = lipgloss.NewStyle().Foreground(styles.AccentColor)
		case "skipped":
			icon = "⊘"
			statusStyle = styles.TextSubtle
		case "pending":
			icon = "○"
		}

		cursor := "  "
		if i == m.diffSectionCursor {
			cursor = "▶ "
		}

		row := fmt.Sprintf("%s%s %s", cursor, statusStyle.Render(icon), styles.TextBold.Render(name))
		if compType != "" {
			row += " " + styles.TextSubtle.Render(fmt.Sprintf("[%s]", compType))
		}
		if isNew {
			row += " " + lipgloss.NewStyle().Foreground(styles.SuccessColor).Render("(new)")
		}
		if cacheStatus == "cache hit" {
			row += " " + styles.TextSubtle.Render("(cached)")
		}

		expanded := m.expandedDiffSections[i]
		if expanded {
			detail := []string{row}
			detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      id:     %s", metaMapStr(build, "component_id"))))
			detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      status: %s", status)))
			if cacheStatus != "" {
				detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      cache:  %s", cacheStatus)))
			}
			if d := metaMapStr(build, "image_digest"); d != "" {
				detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      digest: %s", d)))
			}
			if d := metaMapStr(build, "duration"); d != "" && d != "0" {
				detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      time:   %ss", d)))
			}
			rows = append(rows, strings.Join(detail, "\n"))
		} else {
			rows = append(rows, row)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(1).Margin(0, 1).
		Width(w).
		Render(content)
}

func (m *model) stepDetailViewInstallGroup(meta map[string]interface{}) string {
	w := m.stepDetail.Width() - 4

	installList := toSliceOfMaps(meta["installs"])
	if len(installList) == 0 {
		return ""
	}

	m.diffSectionCount = len(installList)
	if m.diffSectionCursor >= m.diffSectionCount {
		m.diffSectionCursor = 0
	}

	rows := []string{
		styles.TextBold.Render("Install Deployments"),
		styles.TextSubtle.Render("↑/↓ navigate  enter expand  esc collapse"),
		"",
	}

	for i, inst := range installList {
		name := metaMapStr(inst, "install_name")
		status := metaMapStr(inst, "status")
		if name == "" {
			name = metaMapStr(inst, "install_id")
		}

		icon := "○"
		statusStyle := styles.TextSubtle
		switch status {
		case "success":
			icon = "✓"
			statusStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor)
		case "error":
			icon = "✗"
			statusStyle = lipgloss.NewStyle().Foreground(styles.ErrorColor)
		case "in-progress":
			icon = m.spinner.View()
			statusStyle = lipgloss.NewStyle().Foreground(styles.AccentColor)
		}

		cursor := "  "
		if i == m.diffSectionCursor {
			cursor = "▶ "
		}

		row := fmt.Sprintf("%s%s %s", cursor, statusStyle.Render(icon), name)

		expanded := m.expandedDiffSections[i]
		if expanded {
			detail := []string{row}
			detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      install: %s", metaMapStr(inst, "install_id"))))
			if wfID := metaMapStr(inst, "workflow_id"); wfID != "" {
				detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      workflow: %s", wfID)))
			}
			detail = append(detail, styles.TextSubtle.Render(fmt.Sprintf("      status: %s", status)))
			rows = append(rows, strings.Join(detail, "\n"))
		} else {
			rows = append(rows, row)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(1).Margin(0, 1).
		Width(w).
		Render(content)
}

func metaStr(meta map[string]interface{}, key string) string {
	v, ok := meta[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func metaMapStr(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toSliceOfMaps(v interface{}) []map[string]interface{} {
	if v == nil {
		return nil
	}
	switch s := v.(type) {
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(s))
		for _, item := range s {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	case []map[string]interface{}:
		return s
	default:
		return nil
	}
}

func metaMapInt(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}
