package activities

import (
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PRCommentStatus string

const (
	PRCommentStatusPending PRCommentStatus = "pending"
	PRCommentStatusSuccess PRCommentStatus = "success"
	PRCommentStatusFailed  PRCommentStatus = "failed"
	PRCommentStatusSkipped PRCommentStatus = "skipped"
)

// PRCommentPhaseStatus represents the status of one phase check row in the PR comment.
type PRCommentPhaseStatus string

const (
	PRCommentPhaseValidating  PRCommentPhaseStatus = "validating"
	PRCommentPhaseBuilding    PRCommentPhaseStatus = "building"
	PRCommentPhaseConfiguring PRCommentPhaseStatus = "configuring"
	PRCommentPhaseValid       PRCommentPhaseStatus = "valid"
	PRCommentPhaseInvalid     PRCommentPhaseStatus = "invalid"
)

// PRCommentPhases carries the per-phase check statuses rendered in the GitHub PR comment.
// A zero-value Install means the install row is omitted (e.g. build-only mode).
type PRCommentPhases struct {
	Config  PRCommentPhaseStatus
	Builds  PRCommentPhaseStatus
	Install PRCommentPhaseStatus
}

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
	AppBranchID        string
	BranchName         string
	RunID              string
	HeadSHA            string
	RunURL             string
	Status             PRCommentStatus
	Mode               app.AppBranchRunPreviewMode
	Diff               *ComputeAppConfigDiffOutput
	ComponentChanges   []ComponentBuildChange
	InstallImpact      []InstallGroupImpact
	PreviewInstallName string
	PreviewInstallURL  string
	ErrorMessage       string
	// Phases holds per-phase check rows for the GitHub PR comment.  A nil pointer
	// means the Checks section is omitted entirely (e.g. for legacy/skipped runs).
	Phases *PRCommentPhases
	// InstallApplied conditions the "Applied to" block: the section is only
	// rendered when the install step actually succeeded, not just when builds
	// finished with an apply-mode run.
	InstallApplied bool
	// UpdatedAt renders the "Last updated at" line. Workflows leave it zero;
	// the comment activity stamps it at write time so the value reflects when
	// GitHub actually received the edit.
	UpdatedAt time.Time
}

const lastUpdatedPrefix = "**Last updated at:** "

func lastUpdatedLine(t time.Time) string {
	return lastUpdatedPrefix + t.UTC().Format("2006-01-02 15:04:05 MST")
}

// BuildPRCommentBody renders the preview comment. Every run status shares the
// same section skeleton so an edited comment keeps its shape as a run
// progresses; sections appear when their data does, not per status.
func BuildPRCommentBody(p *PRCommentParams) string {
	var b strings.Builder

	if marker := PRCommentMarker(p.AppBranchID); marker != "" {
		b.WriteString(marker + "\n")
	}

	title := fmt.Sprintf("## \U0001f44b Nuon Preview \u2014 %s", previewTitleName(p))
	if label := p.Mode.Label(); label != "" {
		title += fmt.Sprintf(" (%s)", label)
	}
	b.WriteString(title + "\n\n")

	if !p.UpdatedAt.IsZero() {
		b.WriteString(lastUpdatedLine(p.UpdatedAt) + "\n\n")
	}

	if p.RunURL != "" {
		b.WriteString(fmt.Sprintf("[View preview run \u2192](%s)\n\n", p.RunURL))
	} else {
		b.WriteString(fmt.Sprintf("Preview run: `%s`\n\n", p.RunID))
	}

	if hasStackChanges(p) {
		b.WriteString("> [!WARNING]\n")
		b.WriteString("> \U0001f6a8 Stack changes require customers to reprovision the stack. Learn more [here](https://docs.nuon.co/concepts/stacks).\n\n")
	}

	if p.HeadSHA != "" {
		sha := p.HeadSHA
		if len(sha) > 7 {
			sha = sha[:7]
		}
		b.WriteString(fmt.Sprintf("Commit: `%s`\n\n", sha))
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

	if p.Phases != nil {
		writePhaseChecksSection(&b, p.Phases, p.Mode != app.AppBranchRunPreviewModeBuildOnly)
	}

	if note := statusNote(p); note != "" {
		b.WriteString(note + "\n\n")
	}

	if p.Diff != nil {
		writeDiffSection(&b, p.Diff)
	}

	if len(p.ComponentChanges) > 0 {
		writeBuildsSection(&b, p.ComponentChanges)
	}

	if p.Mode == app.AppBranchRunPreviewModeApply {
		if p.InstallApplied && p.PreviewInstallName != "" {
			b.WriteString("### Preview install\n\n")
			if p.PreviewInstallURL != "" {
				b.WriteString(fmt.Sprintf("Applied to [`%s`](%s).\n\n", p.PreviewInstallName, p.PreviewInstallURL))
			} else {
				b.WriteString(fmt.Sprintf("Applied to `%s`.\n\n", p.PreviewInstallName))
			}
		}
	} else if p.Status != PRCommentStatusSkipped &&
		p.Mode != app.AppBranchRunPreviewModeBuildOnly &&
		len(p.InstallImpact) > 0 {
		writeInstallImpactSection(&b, p.InstallImpact)
	}

	if p.ErrorMessage != "" {
		b.WriteString("### Error\n\n")
		b.WriteString(fmt.Sprintf("```\n%s\n```\n\n", p.ErrorMessage))
	}

	b.WriteString("### Debug with MCP\n\n")
	b.WriteString("Copy this prompt into an [MCP-enabled assistant](https://docs.nuon.co/guides/agents/overview):\n\n")
	b.WriteString(fmt.Sprintf("```text\n%s\n```\n", mcpDebugPrompt(p)))

	return b.String()
}

// statusNote is the single progress sentence under the checks table. It tracks
// what the run is doing, while the surrounding sections stay fixed.
func statusNote(p *PRCommentParams) string {
	switch {
	case p.Status == PRCommentStatusSkipped:
		return "No changes to `nuon.toml` detected in this PR. Preview skipped."
	case p.Status == PRCommentStatusPending && p.Diff == nil:
		return "\u23f3 Parsing config..."
	case p.Status == PRCommentStatusPending:
		return "\u23f3 Building components..."
	case p.Status == PRCommentStatusSuccess && p.Mode == app.AppBranchRunPreviewModeBuildOnly:
		return "Builds and config validation succeeded. No install was planned or applied."
	default:
		return ""
	}
}

// writePhaseChecksSection renders the Checks table into the comment body.
// Config is always shown; Builds is always shown; Install is omitted when
// PRCommentPhases.Install is empty.
func writePhaseChecksSection(b *strings.Builder, phases *PRCommentPhases, includeInstall bool) {
	if phases == nil {
		return
	}
	hasInstall := includeInstall && phases.Install != ""
	if phases.Config == "" && phases.Builds == "" && !hasInstall {
		return
	}

	b.WriteString("| Check | Status |\n")
	b.WriteString("|---|---|\n")
	if phases.Config != "" {
		b.WriteString(fmt.Sprintf("| Config | %s |\n", phaseLabel(phases.Config, "config")))
	}
	if phases.Builds != "" {
		b.WriteString(fmt.Sprintf("| Builds | %s |\n", phaseLabel(phases.Builds, "builds")))
	}
	if hasInstall {
		b.WriteString(fmt.Sprintf("| Install | %s |\n", phaseLabel(phases.Install, "install")))
	}
	b.WriteString("\n")
}

// FinalizeFailedPhases converts any still-pending phase rows to Invalid.
// Call this when the run has terminated with failure/cancellation so that
// the PR comment does not show "Validating/Building/Configuring" for phases
// that never completed.
func FinalizeFailedPhases(phases *PRCommentPhases) {
	if phases == nil {
		return
	}
	if isPendingPhase(phases.Config) {
		phases.Config = PRCommentPhaseInvalid
	}
	if isPendingPhase(phases.Builds) {
		phases.Builds = PRCommentPhaseInvalid
	}
	if isPendingPhase(phases.Install) {
		phases.Install = PRCommentPhaseInvalid
	}
}

func isPendingPhase(s PRCommentPhaseStatus) bool {
	return s == PRCommentPhaseValidating || s == PRCommentPhaseBuilding || s == PRCommentPhaseConfiguring
}

func phaseLabel(s PRCommentPhaseStatus, kind string) string {
	switch s {
	case PRCommentPhaseValid:
		return "\u2705 Valid"
	case PRCommentPhaseInvalid:
		return "\u274c Invalid"
	case PRCommentPhaseValidating:
		return "\u23f3 Validating"
	case PRCommentPhaseBuilding:
		return "\u23f3 Building"
	case PRCommentPhaseConfiguring:
		return "\u23f3 Configuring"
	default:
		switch kind {
		case "config":
			return "\u23f3 Validating"
		case "builds":
			return "\u23f3 Building"
		case "install":
			return "\u23f3 Configuring"
		}
		return "\u23f3 In Progress"
	}
}

func mcpDebugPrompt(p *PRCommentParams) string {
	var b strings.Builder
	b.WriteString("Fetch the overview of app branch run ")
	b.WriteString(p.RunID)
	if p.AppName != "" {
		b.WriteString(" for app ")
		b.WriteString(p.AppName)
	}
	if p.BranchName != "" {
		b.WriteString(" branch ")
		b.WriteString(p.BranchName)
	}
	b.WriteString(" and diagnose any failures.")
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
	count := 0
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
		count++
	}
	if rows.Len() == 0 {
		return
	}

	b.WriteString("<details>\n")
	b.WriteString(fmt.Sprintf("<summary><strong>Builds</strong> <code>%d</code></summary>\n\n", count))
	b.WriteString("| Component | Change |\n")
	b.WriteString("|-----------|--------|\n")
	b.WriteString(rows.String())
	b.WriteString("\n</details>\n\n")
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
