package slackrender

import (
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// Message is the rendered output of a Build* call: a fallback text string
// (used by Slack when blocks aren't supported) and the typed Block Kit
// blocks posted as the message body. Blocks are built with the slack-go
// SDK so their shape is validated by the type system rather than
// hand-assembled maps.
type Message struct {
	Text   string
	Blocks []slack.Block
}

// LinkChip is a (label, url) pair rendered either as a context link or an
// actions button.
type LinkChip struct {
	Label string
	URL   string
}

// kv is a single label/value pair rendered as a field in a fielded section.
type kv struct {
	k string
	v string
}

const headerMaxLen = 150

// BuildParentMessage renders the workflow's parent post — a fielded "card"
// emitted on the first event for a workflow. Its shape is identical to
// BuildParentRollup so the post never visually rearranges across edits.
//
// startedAt is the workflow's first-event timestamp. When zero, the
// Duration field is omitted.
func BuildParentMessage(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

// BuildParentRollup renders the parent card on every subsequent edit.
// Identical layout to BuildParentMessage by design.
func BuildParentRollup(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

// buildParent assembles the parent / rollup card: a header, a fielded
// section (status / duration / install / org / by), an optional error
// block, and an "Open in Nuon" button.
func buildParent(e Event, startedAt, now time.Time) Message {
	blocks := []slack.Block{headerBlock(parentHeaderText(e))}

	if fields := parentFields(e, startedAt, now); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if step := parentLatestStep(e); step != "" {
		blocks = append(blocks, slack.NewDividerBlock(), mrkdwnSection(step))
	}
	if errBlock, ok := errorSection(e); ok {
		blocks = append(blocks, errBlock)
	}
	if actions, ok := actionsBlock(buildLinks(e)); ok {
		blocks = append(blocks, actions)
	}

	return Message{Text: plainHeadline(e, true), Blocks: blocks}
}

// parentHeaderText renders the big plain-text header: entity icon +
// workflow title, with the runbook subject appended for runbook runs.
func parentHeaderText(e Event) string {
	parts := nonEmpty(workflowSubjectIcon(e), workflowHeaderTitle(e))
	text := strings.TrimSpace(strings.Join(parts, " "))
	if subj := workflowSubjectLabel(e); subj != "" {
		text = text + " · " + subj
	}
	return text
}

// parentFields builds the two-column field grid. The workflow type is
// intentionally omitted — the header already conveys it.
//
// State is the workflow-level status (In progress → Succeeded / Failed),
// kept distinct from the per-step status so the card never reads as
// "Succeeded" mid-run just because the latest step finished. The latest
// step itself renders as a full-width section below the grid (see
// parentLatestStep) so long step names get room.
func parentFields(e Event, startedAt, now time.Time) []kv {
	fields := []kv{}
	emoji, label := parentState(e)
	fields = append(fields, kv{"State", emoji + " " + label})
	if d := elapsedValue(startedAt, now); d != "" {
		fields = append(fields, kv{"Duration", d})
	}
	if name := installName(e); name != "" {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	if e.Workflow.CreatedByEmail != "" {
		fields = append(fields, kv{"By", slackEscape(e.Workflow.CreatedByEmail)})
	}
	return fields
}

// parentState resolves the workflow-level status for the parent card. Only
// a terminal workflow-kind event yields a terminal state; every step event
// (and the workflow "started") reads as "In progress", because steps only
// flow while the workflow is mid-run.
func parentState(e Event) (string, string) {
	if e.Kind == KindWorkflow {
		switch e.Transition {
		case TransitionSucceeded:
			return statusEmoji(TransitionSucceeded), "Succeeded"
		case TransitionFailed:
			return statusEmoji(TransitionFailed), "Failed"
		case TransitionCancelled:
			return statusEmoji(TransitionCancelled), "Cancelled"
		}
	}
	return "⏳", "In progress"
}

// parentLatestStep renders the full-width "latest step" section line for
// the parent card: a bold step name (which can be long) followed by its
// status on the same line. Empty for workflow-kind events, which carry no
// step.
func parentLatestStep(e Event) string {
	if e.Kind != KindWorkflowStep && e.Kind != KindWorkflowStepApproval {
		return ""
	}
	if e.Step == nil {
		return ""
	}
	title := headerTitle(e)
	if subj := subjectLabel(e); subj != "" && !strings.EqualFold(subj, title) {
		title = title + " — " + subj
	}
	v := "*" + slackEscape(title) + "*  " + statusEmoji(e.Transition) + " " + transitionPhrase(e.Transition)
	if rb := approvalRespondedBy(e); rb != "" {
		v = v + " by " + slackEscape(rb)
	}
	return v
}

// BuildChildMessage renders the per-event threaded reply: a tight two-line
// block. Org / install are intentionally omitted — every reply in the
// thread is the same install, so repeating them is noise; the parent card
// already carries them.
func BuildChildMessage(e Event) Message {
	blocks := []slack.Block{mrkdwnSection(childHeadline(e))}
	if ctx, ok := contextBlock(childContextParts(e), childLinks(e)); ok {
		blocks = append(blocks, ctx)
	}
	return Message{Text: plainHeadline(e, false), Blocks: blocks}
}

// childHeadline renders the reply's lead line: status emoji + bold title +
// optional subject. The status emoji leads so the outcome is scannable
// down the thread gutter.
func childHeadline(e Event) string {
	emoji := statusEmoji(e.Transition)
	title := headerTitle(e)
	headline := strings.TrimSpace(emoji + "  *" + slackEscape(title) + "*")
	if subj := subjectLabel(e); subj != "" && !strings.EqualFold(subj, title) {
		headline = headline + " — " + slackEscape(subj)
	}
	return headline
}

// childContextParts builds the small grey sub-line: status phrase,
// duration, and any error. No org / install (redundant in-thread).
func childContextParts(e Event) []string {
	parts := []string{}
	if phrase := transitionPhrase(e.Transition); phrase != "" {
		if rb := approvalRespondedBy(e); rb != "" {
			phrase = phrase + " by " + rb
		}
		parts = append(parts, phrase)
	}
	if e.Outcome != nil && e.Outcome.DurationMs > 0 {
		parts = append(parts, humanDurationMs(e.Outcome.DurationMs))
	}
	if e.Outcome != nil && e.Outcome.Error != "" {
		parts = append(parts, "error: "+trimContext(e.Outcome.Error, 220))
	}
	return parts
}

// childLinks prefers the most step-specific link (component → sandbox →
// workflow) so the reply jumps the reader to the thing that changed.
func childLinks(e Event) []LinkChip {
	links := []LinkChip{}
	if e.Links == nil {
		return links
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Component },
		func(l *ContextLinks) string { return l.Sandbox },
		func(l *ContextLinks) string { return l.Workflow },
	); l != "" {
		links = append(links, LinkChip{Label: "Open ↗", URL: l})
	}
	if e.Kind == KindWorkflowStepApproval && e.Links.Approval != "" {
		links = append(links, LinkChip{Label: "View approval ↗", URL: e.Links.Approval})
	}
	return links
}

// BuildFlatMessage renders a self-contained top-level message for events
// without a workflow id (e.g. install-created). Unlike a thread reply it
// keeps the org / install context, since it stands alone.
func BuildFlatMessage(e Event) Message {
	blocks := []slack.Block{mrkdwnSection(childHeadline(e))}
	parts := childContextParts(e)
	if name := orgName(e); name != "" {
		parts = append([]string{"org: " + slackEscape(name)}, parts...)
	}
	if name := installName(e); name != "" {
		parts = append([]string{"install: " + slackEscape(name)}, parts...)
	}
	if ctx, ok := contextBlock(parts, buildLinks(e)); ok {
		blocks = append(blocks, ctx)
	}
	return Message{Text: plainHeadline(e, false), Blocks: blocks}
}

// BuildDriftDetectedMessage renders a standalone (non-threaded) drift card.
//
// Drift workflows produce no useful "running drift check" signal for
// subscribers — the only event that matters is "drift was actually
// detected on resource X". So drift events are deliberately NOT anchored
// under a parent post; each detection is its own top-level card that links
// directly to the affected component or sandbox.
func BuildDriftDetectedMessage(e Event) Message {
	blocks := []slack.Block{headerBlock(driftHeaderText(e))}
	if fields := driftFields(e); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(driftLinks(e)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: plainDriftHeadline(e), Blocks: blocks}
}

// driftHeaderText renders the drift card header: "🌊 Drift detected" with
// the component subject appended when resolved.
func driftHeaderText(e Event) string {
	text := "🌊 Drift detected"
	if subject := driftSubject(e); subject != "" {
		text = text + " · " + subject
	}
	return text
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
		return "sandbox"
	}
	return ""
}

// driftFields builds the drift card's field grid: component / install / org.
func driftFields(e Event) []kv {
	fields := []kv{}
	if subject := driftSubject(e); subject != "" {
		fields = append(fields, kv{"Component", slackEscape(subject)})
	}
	if name := installName(e); name != "" {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	return fields
}

// driftLinks returns the drift card button target — the most specific
// resource available (component → sandbox → install).
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

// BuildRoleChangeMessage renders a standalone role-change notification card.
// Role-change events are notification-only signals (like drift) — each lands
// as its own top-level message indicating that a role was enabled or disabled.
func BuildRoleChangeMessage(e Event) Message {
	blocks := []slack.Block{headerBlock(roleChangeHeaderText(e))}
	if fields := roleChangeFields(e); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(roleChangeLinks(e)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: plainRoleChangeHeadline(e), Blocks: blocks}
}

// roleChangeHeaderText renders the header with the role name appended when
// available. Break-glass roles get 🚨, others get 🔐.
func roleChangeHeaderText(e Event) string {
	changeType := roleChangeType(e)
	emoji := "🔐"
	if isBreakGlassRoleType(roleChangeRoleType(e)) {
		emoji = "🚨"
	}
	text := emoji + " Role " + changeType
	if name := roleChangeName(e); name != "" {
		text = text + " · " + name
	}
	return text
}

func isBreakGlassRoleType(roleType string) bool {
	return roleType == "breakglass" || roleType == "runner_breakglass"
}

// roleChangeType extracts the change_type from event metadata. Defaults to
// "changed" when the field is missing.
func roleChangeType(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["change_type"].(string); ok && v != "" {
			return v
		}
	}
	return "changed"
}

// roleChangeName extracts the role_name from event metadata.
func roleChangeName(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["role_name"].(string); ok {
			return v
		}
	}
	return ""
}

// roleChangeRoleType extracts the role_type from event metadata.
func roleChangeRoleType(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["role_type"].(string); ok {
			return v
		}
	}
	return ""
}

// roleChangeFields builds the role-change card's field grid.
func roleChangeFields(e Event) []kv {
	fields := []kv{}
	if name := roleChangeName(e); name != "" {
		fields = append(fields, kv{"Role", slackEscape(name)})
	}
	if roleType := roleChangeRoleType(e); roleType != "" {
		fields = append(fields, kv{"Type", slackEscape(roleType)})
	}
	if name := installName(e); name != "" {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	if names := roleChangeActionTriggers(e); names != "" {
		fields = append(fields, kv{"Action triggers", names})
	}
	return fields
}

func roleChangeActionTriggers(e Event) string {
	if e.Metadata == nil {
		return ""
	}
	if v, ok := e.Metadata["action_trigger_names"].(string); ok && v != "" {
		return slackEscape(v)
	}
	return ""
}

// roleChangeLinks returns the role-change card button targets.
func roleChangeLinks(e Event) []LinkChip {
	links := []LinkChip{}
	if e.Links == nil {
		return links
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Install },
		func(l *ContextLinks) string { return l.Org },
	); l != "" {
		links = append(links, LinkChip{Label: "Open in Nuon", URL: l})
	}
	return links
}

// plainRoleChangeHeadline renders the message-text fallback for role-change events.
func plainRoleChangeHeadline(e Event) string {
	changeType := roleChangeType(e)
	name := roleChangeName(e)
	if name == "" {
		return "🔐 Role " + changeType
	}
	return "🔐 Role " + changeType + " — " + name
}

// workflowSubjectLabel returns a workflow-scoped subject for the parent
// header. For runbook_run workflows that is the runbook's name; other
// workflow types return "" so the header stays the workflow title alone.
func workflowSubjectLabel(e Event) string {
	if e.Workflow.Type == WorkflowTypeRunbookRun && e.Workflow.RunbookName != "" {
		return e.Workflow.RunbookName
	}
	if isAppBranchWorkflow(e) && e.Workflow.OwnerName != "" {
		return e.Workflow.OwnerName
	}
	return ""
}

func isAppBranchWorkflow(e Event) bool {
	switch e.Workflow.Type {
	case WorkflowTypeAppBranchesRun,
		WorkflowTypeAppBranchesConfigRepoUpdate,
		WorkflowTypeAppBranchesComponentRepoUpdate:
		return true
	}
	return e.Workflow.OwnerType == OwnerTypeAppBranches
}

// workflowHeaderTitle returns the title for the parent card — always
// derived from the workflow type, ignoring any step in the event.
func workflowHeaderTitle(e Event) string {
	if title := titleFromWorkflowType(e.Workflow.Type); title != "" {
		return title
	}
	if e.Workflow.Type != "" {
		return e.Workflow.Type
	}
	return "Workflow"
}

// workflowSubjectIcon returns the icon for the parent card — always
// derived from the workflow type, ignoring any step in the event.
func workflowSubjectIcon(e Event) string {
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
	case WorkflowTypeRunbookRun:
		return "📒"
	case WorkflowTypeAppBranchesRun,
		WorkflowTypeAppBranchesConfigRepoUpdate,
		WorkflowTypeAppBranchesComponentRepoUpdate:
		return "🌿"
	}
	return ""
}

// buildLinks resolves the parent card's button(s): "Open in Nuon" +
// (for approvals) "View approval".
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
		links = append(links, LinkChip{Label: "View approval", URL: e.Links.Approval})
	}
	return links
}

// installName resolves the install owner's display name, falling back to a
// truncated id.
func installName(e Event) string {
	if e.Workflow.OwnerType != OwnerTypeInstalls {
		return ""
	}
	if e.Workflow.OwnerName != "" {
		return e.Workflow.OwnerName
	}
	return truncateID(e.Workflow.OwnerID, 10)
}

// orgName resolves the org's display name, falling back to a truncated id.
func orgName(e Event) string {
	if e.OrgName != "" {
		return e.OrgName
	}
	return truncateID(e.OrgID, 10)
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

// subjectLabel resolves the entity name (component name, action name) that
// the event is acting on. The install owner name is intentionally NOT used
// as a fallback — it already appears on the parent card.
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

// approvalRespondedBy returns the responder display string for an approval
// response event, empty otherwise.
func approvalRespondedBy(e Event) string {
	if e.Approval == nil {
		return ""
	}
	return strings.TrimSpace(e.Approval.RespondedBy)
}

// elapsedValue formats the workflow elapsed time as a short phrase, empty
// when startedAt is zero.
func elapsedValue(startedAt, now time.Time) string {
	if startedAt.IsZero() {
		return ""
	}
	if now.IsZero() {
		now = time.Now()
	}
	return humanElapsed(now.Sub(startedAt))
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

// truncateHeader caps header text at Slack's 150-char plain_text limit.
func truncateHeader(s string) string {
	r := []rune(s)
	if len(r) <= headerMaxLen {
		return s
	}
	return string(r[:headerMaxLen-1]) + "…"
}

// firstNonEmptyLink walks a chain of getters and returns the first
// non-empty link.
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

// plainHeadline renders the message-text fallback (used when blocks aren't
// rendered).
func plainHeadline(e Event, parent bool) string {
	parts := []string{}
	if parent {
		if icon := workflowSubjectIcon(e); icon != "" {
			parts = append(parts, icon)
		}
		parts = append(parts, workflowHeaderTitle(e))
		if subj := workflowSubjectLabel(e); subj != "" {
			parts = append(parts, "·", subj)
		}
	} else {
		if icon := subjectIcon(e); icon != "" {
			parts = append(parts, icon)
		}
		parts = append(parts, headerTitle(e))
		if subj := subjectLabel(e); subj != "" {
			parts = append(parts, "—", subj)
		}
	}
	if e.Transition != "" {
		parts = append(parts, "·", transitionPhrase(e.Transition))
	}
	return strings.Join(parts, " ")
}

// subjectIcon returns the entity-kind glyph used by the plain-text
// fallback for step events.
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
	return workflowSubjectIcon(e)
}

// --- Block Kit builders (slack-go typed) ---

// headerBlock builds a large plain_text header block.
func headerBlock(text string) *slack.HeaderBlock {
	return slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, truncateHeader(text), true, false))
}

// mrkdwnSection builds a single-text mrkdwn section block.
func mrkdwnSection(text string) *slack.SectionBlock {
	return slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, text, false, false), nil, nil)
}

// fieldsSection builds a section with a two-column mrkdwn field grid. Each
// pair renders as a bold label above its value.
func fieldsSection(pairs []kv) *slack.SectionBlock {
	fields := make([]*slack.TextBlockObject, 0, len(pairs))
	for _, p := range pairs {
		fields = append(fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+p.k+"*\n"+p.v, false, false))
	}
	return slack.NewSectionBlock(nil, fields, nil)
}

// errorSection builds the full-width error block for terminal failures.
func errorSection(e Event) (*slack.SectionBlock, bool) {
	if e.Outcome == nil || e.Outcome.Error == "" {
		return nil, false
	}
	return mrkdwnSection("*Error*\n```" + trimContext(e.Outcome.Error, 500) + "```"), true
}

// contextBlock builds the small grey context line: footer parts joined by
// " · " followed by link chips. Returns ok=false when empty.
func contextBlock(parts []string, links []LinkChip) (*slack.ContextBlock, bool) {
	elements := []slack.MixedElement{}
	if len(parts) > 0 {
		elements = append(elements, slack.NewTextBlockObject(slack.MarkdownType, slackEscape(strings.Join(parts, " · ")), false, false))
	}
	for _, link := range links {
		if link.URL == "" {
			continue
		}
		elements = append(elements, slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("<%s|%s>", link.URL, link.Label), false, false))
	}
	if len(elements) == 0 {
		return nil, false
	}
	return slack.NewContextBlock("", elements...), true
}

// actionsBlock builds an actions block of URL buttons. URL buttons
// navigate without dispatching an interaction payload, so they need no
// interactivity endpoint. Returns ok=false when there are no links.
func actionsBlock(links []LinkChip) (*slack.ActionBlock, bool) {
	elements := []slack.BlockElement{}
	for _, link := range links {
		if link.URL == "" {
			continue
		}
		btn := slack.NewButtonBlockElement(
			buttonActionID(link.Label),
			"",
			slack.NewTextBlockObject(slack.PlainTextType, link.Label, true, false),
		)
		btn.URL = link.URL
		elements = append(elements, btn)
	}
	if len(elements) == 0 {
		return nil, false
	}
	return slack.NewActionBlock("", elements...), true
}

// buttonActionID derives a stable action_id from a button label. URL
// buttons don't dispatch interactions, but Slack still requires the field.
func buttonActionID(label string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(label) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteByte('_')
		}
	}
	id := strings.Trim(b.String(), "_")
	if id == "" {
		return "action"
	}
	return id
}

// BuildAppConfigSyncedMessage renders a standalone notification card for app
// config sync events. Metadata fields: actor_email, app_name, branch_name.
func BuildAppConfigSyncedMessage(e Event) Message {
	blocks := []slack.Block{headerBlock(appConfigSyncedHeaderText(e))}
	if fields := appConfigSyncedFields(e); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(appConfigSyncedLinks(e)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: plainAppConfigSyncedHeadline(e), Blocks: blocks}
}

func appConfigSyncedHeaderText(e Event) string {
	text := "📦 App config synced"
	if name := metadataString(e, "app_name"); name != "" {
		text = text + " · " + name
	}
	return text
}

func appConfigSyncedFields(e Event) []kv {
	fields := []kv{}
	if name := metadataString(e, "app_name"); name != "" {
		fields = append(fields, kv{"App", slackEscape(name)})
	}
	if name := metadataString(e, "branch_name"); name != "" {
		fields = append(fields, kv{"Branch", slackEscape(name)})
	}
	if actor := metadataString(e, "actor_email"); actor != "" {
		fields = append(fields, kv{"By", slackEscape(actor)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	return fields
}

func appConfigSyncedLinks(e Event) []LinkChip {
	if e.Links == nil {
		return nil
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Org },
	); l != "" {
		return []LinkChip{{Label: "Open in Nuon", URL: l}}
	}
	return nil
}

func plainAppConfigSyncedHeadline(e Event) string {
	text := "📦 App config synced"
	if name := metadataString(e, "app_name"); name != "" {
		text = text + " — " + name
	}
	return text
}

func metadataString(e Event, key string) string {
	if e.Metadata == nil {
		return ""
	}
	if v, ok := e.Metadata[key].(string); ok {
		return v
	}
	return ""
}

// slackEscape escapes the three characters Slack treats specially in
// mrkdwn (<, >, &). Newlines pass through.
func slackEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
