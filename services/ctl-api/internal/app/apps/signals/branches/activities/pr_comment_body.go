package activities

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PRCommentStatus string

const (
	PRCommentStatusPending PRCommentStatus = "pending"
	PRCommentStatusSuccess PRCommentStatus = "success"
	PRCommentStatusFailed  PRCommentStatus = "failed"
	PRCommentStatusSkipped PRCommentStatus = "skipped"
)

// InstallImpact is what a preview run would change on a single install if the
// config were applied. Nothing is applied to produce it.
type InstallImpact struct {
	InstallID      string
	InstallName    string
	Added          int
	Changed        int
	Removed        int
	Unchanged      int
	SandboxChanged bool
	StackChanged   bool
}

type InstallGroupImpact struct {
	GroupName string
	Installs  []InstallImpact
}

type ComponentBuildChange struct {
	ComponentName string `json:"component_name"`
	ComponentID   string `json:"component_id"`
	BuildID       string `json:"build_id"`
	ChangeReason  string `json:"change_reason"`
	BuildURL      string `json:"build_url"`
}

type PRCommentParams struct {
	OrgName            string
	AppName            string
	BranchName         string
	RunID              string
	RunURL             string
	Status             PRCommentStatus
	Mode               app.AppBranchRunPreviewMode
	Diff               *ComputeAppConfigDiffOutput
	ComponentChanges   []ComponentBuildChange
	InstallImpact      []InstallGroupImpact
	PreviewInstallName string
	PreviewInstallURL  string
	ErrorMessage       string
}

func BuildPRCommentBody(p *PRCommentParams) string {
	var b strings.Builder

	title := fmt.Sprintf("## Nuon Preview \u2014 %s", previewTitleName(p))
	if label := p.Mode.Label(); label != "" {
		title += fmt.Sprintf(" (%s)", label)
	}
	b.WriteString(title + "\n\n")

	if p.RunURL != "" {
		b.WriteString(fmt.Sprintf("[View preview run \u2192](%s)\n\n", p.RunURL))
	} else {
		b.WriteString(fmt.Sprintf("Preview run: `%s`\n\n", p.RunID))
	}

	if hasStackChanges(p) {
		b.WriteString("> [!WARNING]\n")
		b.WriteString("> \U0001f6a8 Stack changes require customers to reprovision the stack. Learn more [here](https://docs.nuon.co/concepts/stacks).\n\n")
	}

	switch p.Status {
	case PRCommentStatusPending:
		b.WriteString("**Status**: \u23f3 In Progress\n\n")
	case PRCommentStatusSuccess:
		b.WriteString("**Status**: \u2705 Complete\n\n")
	case PRCommentStatusFailed:
		b.WriteString("**Status**: \u274c Failed\n\n")
	case PRCommentStatusSkipped:
		b.WriteString("**Status**: \u2298 No Changes\n\n")
	}

	if p.Status == PRCommentStatusSkipped {
		b.WriteString("No changes to `nuon.toml` detected in this PR. Preview skipped.\n")
	} else if p.Status == PRCommentStatusFailed && p.ErrorMessage != "" {
		if p.Diff != nil {
			writeDiffSection(&b, p.Diff)
		}
		b.WriteString("### Error\n\n")
		b.WriteString(fmt.Sprintf("```\n%s\n```\n", p.ErrorMessage))
	} else if p.Status == PRCommentStatusPending {
		if p.Diff != nil {
			writeDiffSection(&b, p.Diff)
			b.WriteString("\u23f3 Building components...\n")
		} else {
			b.WriteString("\u23f3 Parsing config...\n")
		}
	} else if p.Diff != nil {
		writeDiffSection(&b, p.Diff)
	}

	if p.Status != PRCommentStatusPending && p.Status != PRCommentStatusSkipped && len(p.ComponentChanges) > 0 {
		writeBuildsSection(&b, p.ComponentChanges)
	}

	if p.Status == PRCommentStatusSuccess && p.Mode == app.AppBranchRunPreviewModeBuildOnly {
		b.WriteString("Builds and config validation succeeded. No install was planned or applied.\n")
	}

	if p.Status == PRCommentStatusSuccess && p.Mode == app.AppBranchRunPreviewModeApply && p.PreviewInstallName != "" {
		b.WriteString("### Preview install\n\n")
		if p.PreviewInstallURL != "" {
			b.WriteString(fmt.Sprintf("Applied to [`%s`](%s).\n", p.PreviewInstallName, p.PreviewInstallURL))
		} else {
			b.WriteString(fmt.Sprintf("Applied to `%s`.\n", p.PreviewInstallName))
		}
	} else if p.Status != PRCommentStatusSkipped &&
		p.Mode != app.AppBranchRunPreviewModeBuildOnly &&
		p.Mode != app.AppBranchRunPreviewModeApply &&
		len(p.InstallImpact) > 0 {
		writeInstallImpactSection(&b, p.InstallImpact)
	}

	if p.Status != PRCommentStatusSkipped {
		b.WriteString("\n### Debug with MCP\n\n")
		b.WriteString("Copy this prompt into an MCP-enabled assistant:\n\n")
		b.WriteString(fmt.Sprintf("```text\nFetch the overview of app branch run %s and diagnose any failures.\n```\n", p.RunID))
	}

	return b.String()
}

func previewTitleName(p *PRCommentParams) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{p.OrgName, p.AppName, p.BranchName} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "App"
	}
	return strings.Join(parts, "/")
}

func hasStackChanges(p *PRCommentParams) bool {
	if p.Diff != nil {
		for _, section := range p.Diff.Sections {
			if strings.EqualFold(section.Name, "stack") &&
				(section.Additions > 0 || section.Changed > 0 || section.Removals > 0) {
				return true
			}
		}
	}
	for _, group := range p.InstallImpact {
		for _, install := range group.Installs {
			if install.StackChanged {
				return true
			}
		}
	}
	return false
}

func writeBuildsSection(b *strings.Builder, changes []ComponentBuildChange) {
	var rows strings.Builder
	for _, change := range changes {
		label := ""
		switch change.ChangeReason {
		case ChangeReasonSourceChanged:
			label = "Source changed"
		case ChangeReasonConfigChanged:
			label = "Config changed"
		default:
			continue
		}
		component := fmt.Sprintf("`%s`", change.ComponentName)
		if change.BuildURL != "" {
			component = fmt.Sprintf("[`%s`](%s)", change.ComponentName, change.BuildURL)
		}
		rows.WriteString(fmt.Sprintf("| %s | `%s` |\n", component, label))
	}
	if rows.Len() == 0 {
		return
	}

	b.WriteString("### Builds\n\n")
	b.WriteString("| Component | Change |\n")
	b.WriteString("|-----------|--------|\n")
	b.WriteString(rows.String())
	b.WriteString("\n")
}

// writeDiffSection mirrors the dashboard overview's config diff card: a single
// collapsed disclosure whose summary carries the aggregate counts, expanding to
// operation-prefixed entity rows grouped by section.
func writeDiffSection(b *strings.Builder, diff *ComputeAppConfigDiffOutput) {
	var added, changed, removed int
	for _, s := range diff.Sections {
		added += s.Additions
		changed += s.Changed
		removed += s.Removals
	}

	b.WriteString("<details>\n")
	b.WriteString(fmt.Sprintf("<summary><strong>Config changes</strong> %s</summary>\n\n",
		diffCountSummary(added, changed, removed)))

	if len(diff.Sections) == 0 {
		b.WriteString("No config changes.\n\n</details>\n\n")
		return
	}

	for _, s := range diff.Sections {
		b.WriteString(fmt.Sprintf("#### %s\n\n", s.Name))

		if len(s.Entries) == 0 {
			b.WriteString(diffCountSummary(s.Additions, s.Changed, s.Removals) + "\n\n")
			continue
		}

		for _, e := range s.Entries {
			row := fmt.Sprintf("- `%s` `%s`", diffOpSymbol(e.Op), e.Name)
			if e.Description != "" {
				row += fmt.Sprintf(" \u2014 %s", e.Description)
			}
			b.WriteString(row + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("</details>\n\n")
}

// Counts are wrapped in literal <code> rather than backticks because GitHub
// does not render markdown inside a <summary>.
func diffCountSummary(added, changed, removed int) string {
	parts := make([]string, 0, 3)
	if added > 0 {
		parts = append(parts, fmt.Sprintf("<code>+%d</code>", added))
	}
	if changed > 0 {
		parts = append(parts, fmt.Sprintf("<code>~%d</code>", changed))
	}
	if removed > 0 {
		parts = append(parts, fmt.Sprintf("<code>-%d</code>", removed))
	}
	if len(parts) == 0 {
		return "<code>no changes</code>"
	}
	return strings.Join(parts, " ")
}

func diffOpSymbol(op string) string {
	switch op {
	case "add":
		return "+"
	case "remove":
		return "-"
	case "change":
		return "~"
	default:
		return "\u2022"
	}
}

func writeInstallImpactSection(b *strings.Builder, groups []InstallGroupImpact) {
	total := 0
	for _, g := range groups {
		total += len(g.Installs)
	}

	b.WriteString(fmt.Sprintf("### Install Impact — %d install(s)\n\n", total))
	b.WriteString("Preview only — nothing was applied to these installs.\n\n")

	for _, g := range groups {
		if len(g.Installs) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("<details><summary><b>%s</b> (%d)</summary>\n\n", g.GroupName, len(g.Installs)))
		b.WriteString("| Install | Added | Changed | Removed | Sandbox | Stack |\n")
		b.WriteString("|---------|-------|---------|---------|---------|-------|\n")
		for _, i := range g.Installs {
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				i.InstallName,
				formatCount(i.Added, "+"),
				formatCount(i.Changed, ""),
				formatCount(i.Removed, ""),
				formatChanged(i.SandboxChanged),
				formatChanged(i.StackChanged),
			))
		}
		b.WriteString("\n</details>\n\n")
	}
}

func formatChanged(changed bool) string {
	if changed {
		return "⚠️ changed"
	}
	return "—"
}

func formatCount(n int, prefix string) string {
	if n == 0 {
		return "0"
	}
	return fmt.Sprintf("%s%d", prefix, n)
}
