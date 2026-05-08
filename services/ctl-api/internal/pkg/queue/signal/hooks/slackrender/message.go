package slackrender

import (
	"fmt"
	"strings"
	"time"
)

// Message is the rendered output of a Build* call: a fallback text string
// (used by Slack when blocks aren't supported) and the structured blocks
// posted as the message body.
type Message struct {
	Text   string
	Blocks []any
}

// LinkChip is a (label, url) pair rendered as a clickable mrkdwn chip in
// the trailing context block.
type LinkChip struct {
	Label string
	URL   string
}

// BuildParentMessage renders the workflow's parent post — emitted on the
// first event for a workflow. The parent's shape (headline + status +
// footer + links) is identical to BuildParentRollup so the post never
// visually grows or rearranges across edits.
//
// startedAt is the workflow's first-event timestamp. When zero, the
// "running 5m" / "ran 5m" line is omitted.
func BuildParentMessage(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

// BuildParentRollup renders the parent post on every subsequent edit.
// Identical layout to BuildParentMessage by design.
func BuildParentRollup(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

// BuildChildMessage renders the per-event threaded reply.
func BuildChildMessage(e Event) Message {
	headline := childHeadline(e)
	statusLine := childStatusLine(e)
	footer := childFooterParts(e)
	links := buildLinks(e)

	return Message{
		Text:   plainHeadline(e, false),
		Blocks: buildBlocks(headline, statusLine, footer, links),
	}
}

// BuildFlatMessage renders a defensive single-block fallback for events
// without a workflow id (e.g. install-created). Same shape as the child.
func BuildFlatMessage(e Event) Message {
	return BuildChildMessage(e)
}

// BuildDriftDetectedMessage renders a standalone (non-threaded) drift
// notification.
//
// Drift workflows produce no useful "running drift check" signal for
// subscribers — the only event that matters is "drift was actually
// detected on resource X". So drift events are deliberately NOT anchored
// under a parent post; each detection is its own top-level message that
// links directly to the affected component or sandbox.
//
// Headline: "🌊 Drift detected — <component name>" for component drift,
// "🌊 Drift detected on sandbox" for sandbox drift. Footer carries the
// org / install chips. The link chip points at the most specific
// resource available (component → sandbox → install).
func BuildDriftDetectedMessage(e Event) Message {
	headline := driftHeadline(e)
	footer := driftFooterParts(e)
	links := driftLinks(e)

	return Message{
		Text:   plainDriftHeadline(e),
		Blocks: buildBlocks(headline, "", footer, links),
	}
}

// driftHeadline renders the bold section line for a drift-detected event.
func driftHeadline(e Event) string {
	subject := driftSubject(e)
	headline := "🌊 *Drift detected*"
	if subject != "" {
		headline = headline + " — " + slackEscape(subject)
	}
	return headline
}

// driftSubject resolves the subject (component name or sandbox literal)
// from the enriched step. Empty when the step couldn't be resolved.
func driftSubject(e Event) string {
	if e.Step == nil {
		return ""
	}
	switch e.Step.TargetType {
	case TargetTypeInstallDeploys:
		return e.Step.ComponentName
	case TargetTypeInstallSandboxRuns:
		// There's no SandboxName on the step ref (only id), and there's
		// one sandbox per install, so the install chip in the footer is
		// already enough disambiguation. Use a generic "sandbox" label.
		return "sandbox"
	}
	return ""
}

// driftFooterParts builds the small grey context block for drift events:
// org and install chips only. Workflow / elapsed / status are
// intentionally omitted — drift events are point-in-time observations,
// not lifecycle transitions, so the usual "running 5m" / status emoji
// content would be noise.
func driftFooterParts(e Event) []string {
	parts := []string{}
	if e.OrgName != "" {
		parts = append(parts, "org: "+e.OrgName)
	} else if e.OrgID != "" {
		parts = append(parts, "org: "+truncateID(e.OrgID, 10))
	}
	if e.Workflow.OwnerType == OwnerTypeInstalls && e.Workflow.OwnerName != "" {
		parts = append(parts, "install: "+e.Workflow.OwnerName)
	} else if e.Workflow.OwnerType == OwnerTypeInstalls && e.Workflow.OwnerID != "" {
		parts = append(parts, "install: "+truncateID(e.Workflow.OwnerID, 10))
	}
	return parts
}

// driftLinks returns the link chips for a drift event. Prefer the most
// specific resource link — component → sandbox → install — so the click
// lands the reader directly on the thing that drifted.
func driftLinks(e Event) []LinkChip {
	links := []LinkChip{}
	if e.Links == nil {
		return links
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Component },
		func(l *ContextLinks) string { return l.Sandbox },
		func(l *ContextLinks) string { return l.Install },
		func(l *ContextLinks) string { return l.Workflow },
		func(l *ContextLinks) string { return l.Org },
	); l != "" {
		links = append(links, LinkChip{Label: "Open in Nuon", URL: l})
	}
	return links
}

// plainDriftHeadline renders the message-text fallback for drift events.
func plainDriftHeadline(e Event) string {
	subject := driftSubject(e)
	if subject == "" {
		return "🌊 Drift detected"
	}
	return "🌊 Drift detected — " + subject
}

// buildParent assembles the parent / rollup blocks.
func buildParent(e Event, startedAt, now time.Time) Message {
	headline := parentHeadline(e)
	statusLine := parentStatusLine(e, startedAt, now)
	footer := parentFooterParts(e, startedAt, now)
	links := buildLinks(e)

	return Message{
		Text:   plainHeadline(e, true),
		Blocks: buildBlocks(headline, statusLine, footer, links),
	}
}

// parentHeadline renders the bolded section line — entity icon +
// sentence-case title. The install/owner name and workflow type are
// intentionally NOT appended here: both already appear as chips in the
// context block below, and inlining them turns the headline into noise
// like "Workflow step — bdoans _(provision)_".
func parentHeadline(e Event) string {
	icon := subjectIcon(e)
	title := headerTitle(e)
	headline := strings.TrimSpace(strings.Join(nonEmpty(icon, title), " "))
	headline = "*" + slackEscape(headline) + "*"
	if subj := subjectLabel(e); subj != "" && !strings.EqualFold(subj, title) {
		headline = headline + " — " + slackEscape(subj)
	}
	return headline
}

// parentStatusLine renders the second line in the section block: the
// freshest transition. The step name is intentionally NOT repeated here
// because it is now the headline.
func parentStatusLine(e Event, startedAt, now time.Time) string {
	if e.Transition == "" {
		return ""
	}
	parts := []string{"latest:", statusPart(e)}
	if rb := approvalRespondedBy(e); rb != "" {
		parts = append(parts, "by "+slackEscape(rb))
	}
	return strings.Join(parts, " ")
}

// parentFooterParts builds the small grey context block: org / install /
// workflow chips, then a "running 5m" / "ran 5m" elapsed line, then any
// error.
func parentFooterParts(e Event, startedAt, now time.Time) []string {
	parts := []string{}
	if e.OrgName != "" {
		parts = append(parts, "org: "+e.OrgName)
	} else if e.OrgID != "" {
		parts = append(parts, "org: "+truncateID(e.OrgID, 10))
	}
	if e.Workflow.OwnerType == OwnerTypeInstalls && e.Workflow.OwnerName != "" {
		parts = append(parts, "install: "+e.Workflow.OwnerName)
	} else if e.Workflow.OwnerType == OwnerTypeInstalls && e.Workflow.OwnerID != "" {
		parts = append(parts, "install: "+truncateID(e.Workflow.OwnerID, 10))
	}
	if e.Workflow.Type != "" {
		parts = append(parts, "workflow: "+e.Workflow.Type)
	}
	if elapsed := buildElapsed(e, startedAt, now); elapsed != "" {
		parts = append(parts, elapsed)
	}
	if e.Outcome != nil && e.Outcome.Error != "" {
		parts = append(parts, "error: "+trimContext(e.Outcome.Error, 220))
	}
	return parts
}

// childHeadline renders a child's bolded section line.
func childHeadline(e Event) string {
	icon := subjectIcon(e)
	title := headerTitle(e)
	headline := strings.TrimSpace(strings.Join(nonEmpty(icon, title), " "))
	if subj := subjectLabel(e); subj != "" && !strings.EqualFold(subj, title) {
		headline = headline + " — " + subj
	}
	return "*" + slackEscape(headline) + "*"
}

// childStatusLine renders the per-event status sub-line. The step name
// is intentionally NOT repeated here because it is now the headline.
func childStatusLine(e Event) string {
	if e.Transition == "" {
		return ""
	}
	parts := []string{statusPart(e)}
	if rb := approvalRespondedBy(e); rb != "" {
		parts = append(parts, "by "+slackEscape(rb))
	}
	return strings.Join(parts, " · ")
}

// childFooterParts builds the small grey context block for child posts.
func childFooterParts(e Event) []string {
	parts := []string{}
	if e.OrgName != "" {
		parts = append(parts, "org: "+e.OrgName)
	}
	if e.Workflow.OwnerType == OwnerTypeInstalls && e.Workflow.OwnerName != "" {
		parts = append(parts, "install: "+e.Workflow.OwnerName)
	}
	if e.Outcome != nil && e.Outcome.DurationMs > 0 {
		parts = append(parts, humanDurationMs(e.Outcome.DurationMs))
	}
	if e.Outcome != nil && e.Outcome.Error != "" {
		parts = append(parts, "error: "+trimContext(e.Outcome.Error, 220))
	}
	return parts
}

// buildLinks resolves the right-side link chips for the context block:
// "Open in Nuon" + (for approvals) "view approval".
func buildLinks(e Event) []LinkChip {
	links := []LinkChip{}
	if e.Links == nil {
		return links
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Workflow },
		func(l *ContextLinks) string { return l.Install },
		func(l *ContextLinks) string { return l.Org },
	); l != "" {
		links = append(links, LinkChip{Label: "Open in Nuon", URL: l})
	}
	if e.Kind == KindWorkflowStepApproval && e.Links.Approval != "" {
		links = append(links, LinkChip{Label: "view approval", URL: e.Links.Approval})
	}
	return links
}

// statusPart renders the "<emoji> <phrase>" status segment used by every
// per-event line.
func statusPart(e Event) string {
	return fmt.Sprintf("%s %s", statusEmoji(e.Transition), transitionPhrase(e.Transition))
}

// statusEmoji maps the transition to a single-glyph prefix.
func statusEmoji(transition string) string {
	switch strings.ToLower(strings.TrimSpace(transition)) {
	case TransitionStarted:
		return "⏳"
	case TransitionSucceeded:
		return "✅"
	case TransitionFailed:
		return "❌"
	case TransitionCancelled:
		return "🚫"
	case TransitionRequested:
		return "⏸️"
	case TransitionApproved:
		return "👍"
	case TransitionRejected:
		return "👎"
	default:
		return "▫️"
	}
}

// subjectIcon returns the entity-kind glyph that prefixes the headline.
func subjectIcon(e Event) string {
	if (e.Kind == KindWorkflowStep || e.Kind == KindWorkflowStepApproval) && e.Step != nil {
		switch e.Step.TargetType {
		case TargetTypeInstallDeploys:
			return "🧩"
		case TargetTypeInstallSandboxRuns:
			return "📦"
		case TargetTypeInstallActionWorkflowRuns:
			return "🏃"
		case TargetTypeInstallCloudFormationStack,
			TargetTypeInstallRunnerUpdate:
			return "🏗"
		}
	}
	switch e.Workflow.Type {
	case WorkflowTypeProvision,
		WorkflowTypeReprovision,
		WorkflowTypeDeprovision,
		WorkflowTypeInputUpdate,
		WorkflowTypeSyncSecrets:
		return "🏗"
	case WorkflowTypeDeprovisionSandbox,
		WorkflowTypeReprovisionSandbox:
		return "📦"
	case WorkflowTypeManualDeploy,
		WorkflowTypeDeployComponents,
		WorkflowTypeTeardownComponent,
		WorkflowTypeTeardownComponents,
		WorkflowTypeDriftRun:
		return "🧩"
	case WorkflowTypeActionWorkflowRun:
		return "🏃"
	}
	return ""
}

// subjectLabel resolves the entity name (component name, action name)
// that the event is acting on. Used to suffix the headline so a reader
// sees "Deploying component — api" instead of just "Deploying
// component". The install owner name is intentionally NOT used as a
// fallback — it already appears as a chip in the context block, and
// inlining it just made the headline noisy ("Workflow step — bdoans").
func subjectLabel(e Event) string {
	if (e.Kind == KindWorkflowStep || e.Kind == KindWorkflowStepApproval) && e.Step != nil {
		switch e.Step.TargetType {
		case TargetTypeInstallDeploys:
			if e.Step.ComponentName != "" {
				return e.Step.ComponentName
			}
			return ""
		case TargetTypeInstallSandboxRuns:
			return ""
		case TargetTypeInstallActionWorkflowRuns:
			if e.Parent != nil && e.Parent.ActionName != "" {
				return e.Parent.ActionName
			}
			return ""
		}
	}
	return ""
}

// approvalRespondedBy returns the responder display string for an
// approval response event, empty otherwise.
func approvalRespondedBy(e Event) string {
	if e.Approval == nil {
		return ""
	}
	return strings.TrimSpace(e.Approval.RespondedBy)
}

// buildElapsed produces a "running 5m" / "ran 5m" segment. Returns ""
// when startedAt is zero.
func buildElapsed(e Event, startedAt, now time.Time) string {
	if startedAt.IsZero() {
		return ""
	}
	if now.IsZero() {
		now = time.Now()
	}
	verb := "running"
	if e.IsTerminal() {
		verb = "ran"
	}
	return fmt.Sprintf("%s %s", verb, humanElapsed(now.Sub(startedAt)))
}

// humanElapsed formats a duration as a short human-readable phrase.
func humanElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return d.Round(time.Second).String()
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Round(time.Minute)/time.Minute))
	case d < 24*time.Hour:
		d = d.Round(time.Minute)
		hours := int(d / time.Hour)
		mins := int((d % time.Hour) / time.Minute)
		if mins == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		d = d.Round(time.Hour)
		days := int(d / (24 * time.Hour))
		hours := int((d % (24 * time.Hour)) / time.Hour)
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

// humanDurationMs converts a millisecond duration to a short string.
func humanDurationMs(durationMs int64) string {
	d := time.Duration(durationMs) * time.Millisecond
	if d < time.Second {
		return d.String()
	}
	return d.Round(time.Second).String()
}

// trimContext caps a context-line value at maxLen runes, replacing
// newlines with spaces and appending an ellipsis when truncated.
func trimContext(raw string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	cleaned := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
	if len(cleaned) <= maxLen {
		return cleaned
	}
	if maxLen <= 3 {
		return cleaned[:maxLen]
	}
	return cleaned[:maxLen-3] + "..."
}

// truncateID renders a ULID-ish id as the leading n runes followed by an
// ellipsis when longer than n.
func truncateID(id string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(id)
	if len(r) <= n {
		return id
	}
	return string(r[:n]) + "…"
}

// firstNonEmptyLink walks a chain of getters and returns the first non-empty
// link.
func firstNonEmptyLink(links *ContextLinks, getters ...func(*ContextLinks) string) string {
	if links == nil {
		return ""
	}
	for _, g := range getters {
		if v := g(links); v != "" {
			return v
		}
	}
	return ""
}

// nonEmpty filters out empty strings.
func nonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// plainHeadline renders the message-text fallback (used when blocks
// aren't rendered).
func plainHeadline(e Event, parent bool) string {
	parts := []string{}
	if icon := subjectIcon(e); icon != "" {
		parts = append(parts, icon)
	}
	parts = append(parts, headerTitle(e))
	if subj := subjectLabel(e); subj != "" {
		parts = append(parts, "—", subj)
	}
	if e.Transition != "" {
		parts = append(parts, "·", transitionPhrase(e.Transition))
	}
	return strings.Join(parts, " ")
}

// buildBlocks constructs the canonical two-block Slack message body:
//
//	SECTION: bold headline
//	  (optional second mrkdwn line: status / latest line)
//	CONTEXT: small grey footer parts joined by " · " + link chips
//
// All link chips render as separate context elements so Slack groups them
// inline (slackbot inlines everything; we keep them split — see ctl-api
// AGENTS.md for the rationale).
func buildBlocks(headlineLine, statusLine string, footerParts []string, links []LinkChip) []any {
	sectionText := headlineLine
	if statusLine != "" {
		sectionText = sectionText + "\n" + statusLine
	}
	blocks := []any{
		map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": sectionText,
			},
		},
	}

	contextElements := []any{}
	if len(footerParts) > 0 {
		contextElements = append(contextElements, map[string]any{
			"type": "mrkdwn",
			"text": slackEscape(strings.Join(footerParts, " · ")),
		})
	}
	for _, link := range links {
		if link.URL == "" {
			continue
		}
		contextElements = append(contextElements, map[string]any{
			"type": "mrkdwn",
			"text": fmt.Sprintf("<%s|%s>", link.URL, link.Label),
		})
	}
	if len(contextElements) > 0 {
		blocks = append(blocks, map[string]any{
			"type":     "context",
			"elements": contextElements,
		})
	}
	return blocks
}

// slackEscape escapes the three characters Slack treats specially in
// mrkdwn (<, >, &). Newlines pass through.
func slackEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
