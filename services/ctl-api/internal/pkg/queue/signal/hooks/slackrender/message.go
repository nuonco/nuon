package slackrender

import (
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type Message struct {
	Text   string
	Blocks []slack.Block
}

type LinkChip struct {
	Label string
	URL   string
}

type kv struct {
	k string
	v string
}

const headerMaxLen = 150

func BuildParentMessage(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

func BuildParentRollup(e Event, startedAt time.Time) Message {
	return buildParent(e, startedAt, time.Now())
}

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

func parentHeaderText(e Event) string {
	parts := nonEmpty(workflowSubjectIcon(e), workflowHeaderTitle(e))
	text := strings.TrimSpace(strings.Join(parts, " "))
	if subj := workflowSubjectLabel(e); subj != "" {
		text = text + " · " + subj
	}
	return text
}

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

func BuildChildMessage(e Event) Message {
	blocks := []slack.Block{mrkdwnSection(childHeadline(e))}
	if ctx, ok := contextBlock(childContextParts(e), childLinks(e)); ok {
		blocks = append(blocks, ctx)
	}
	return Message{Text: plainHeadline(e, false), Blocks: blocks}
}

func childHeadline(e Event) string {
	emoji := statusEmoji(e.Transition)
	title := headerTitle(e)
	headline := strings.TrimSpace(emoji + "  *" + slackEscape(title) + "*")
	if subj := subjectLabel(e); subj != "" && !strings.EqualFold(subj, title) {
		headline = headline + " — " + slackEscape(subj)
	}
	return headline
}

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

func driftHeaderText(e Event) string {
	text := "🌊 Drift detected"
	if subject := driftSubject(e); subject != "" {
		text = text + " · " + subject
	}
	return text
}

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

func plainDriftHeadline(e Event) string {
	subject := driftSubject(e)
	if subject == "" {
		return "🌊 Drift detected"
	}
	return "🌊 Drift detected — " + subject
}

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

func roleChangeType(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["change_type"].(string); ok && v != "" {
			return v
		}
	}
	return "changed"
}

func roleChangeName(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["role_name"].(string); ok {
			return v
		}
	}
	return ""
}

func roleChangeRoleType(e Event) string {
	if e.Metadata != nil {
		if v, ok := e.Metadata["role_type"].(string); ok {
			return v
		}
	}
	return ""
}

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

func plainRoleChangeHeadline(e Event) string {
	changeType := roleChangeType(e)
	name := roleChangeName(e)
	if name == "" {
		return "🔐 Role " + changeType
	}
	return "🔐 Role " + changeType + " — " + name
}

func BuildRunnerUnhealthyMessage(e Event) Message {
	blocks := []slack.Block{headerBlock(runnerUnhealthyHeaderText(e))}
	if fields := runnerUnhealthyFields(e); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(runnerUnhealthyLinks(e)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: plainRunnerUnhealthyHeadline(e), Blocks: blocks}
}

func runnerUnhealthyHeaderText(e Event) string {
	text := "🚨 Runner unhealthy"
	if name := runnerUnhealthyName(e); name != "" {
		text += " · " + name
	}
	return text
}

func runnerUnhealthyName(e Event) string {
	if name := metadataString(e, "runner_name"); name != "" {
		return name
	}
	return truncateID(metadataString(e, "runner_id"), 10)
}

func runnerUnhealthyFields(e Event) []kv {
	fields := []kv{}
	if name := runnerUnhealthyName(e); name != "" {
		fields = append(fields, kv{"Runner", slackEscape(name)})
	}
	if groupType := metadataString(e, "runner_group_type"); groupType != "" {
		fields = append(fields, kv{"Scope", slackEscape(groupType)})
	}
	if name := installName(e); name != "" {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	fromStatus := metadataString(e, "from_status")
	toStatus := metadataString(e, "to_status")
	if fromStatus != "" && toStatus != "" {
		fields = append(fields, kv{"Status", slackEscape(fromStatus + " → " + toStatus)})
	}
	if reason := metadataString(e, "reason"); reason != "" {
		fields = append(fields, kv{"Reason", slackEscape(reason)})
	}
	return fields
}

func runnerUnhealthyLinks(e Event) []LinkChip {
	if e.Links == nil {
		return nil
	}
	if link := firstNonEmptyLink(e.Links,
		func(links *ContextLinks) string { return links.Install },
		func(links *ContextLinks) string { return links.Org },
	); link != "" {
		return []LinkChip{{Label: "Open in Nuon", URL: link}}
	}
	return nil
}

func plainRunnerUnhealthyHeadline(e Event) string {
	text := "🚨 Runner unhealthy"
	if name := runnerUnhealthyName(e); name != "" {
		text += " — " + name
	}
	return text
}

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

func workflowHeaderTitle(e Event) string {
	if title := titleFromWorkflowType(e.Workflow.Type); title != "" {
		return title
	}
	if e.Workflow.Type != "" {
		return e.Workflow.Type
	}
	return "Workflow"
}

func workflowSubjectIcon(e Event) string {
	switch e.Workflow.Type {
	case WorkflowTypeProvision,
		WorkflowTypeReprovision,
		WorkflowTypeReprovisionStack,
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
		WorkflowTypeDriftRun,
		WorkflowTypeComponentEnabled,
		WorkflowTypeComponentDisabled:
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

func installName(e Event) string {
	if e.Workflow.OwnerType != OwnerTypeInstalls {
		return ""
	}
	if e.Workflow.OwnerName != "" {
		return e.Workflow.OwnerName
	}
	return truncateID(e.Workflow.OwnerID, 10)
}

func orgName(e Event) string {
	if e.OrgName != "" {
		return e.OrgName
	}
	return truncateID(e.OrgID, 10)
}

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
	case TransitionAwaitingRetry:
		return "⚠️"
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

func approvalRespondedBy(e Event) string {
	if e.Approval == nil {
		return ""
	}
	return strings.TrimSpace(e.Approval.RespondedBy)
}

func elapsedValue(startedAt, now time.Time) string {
	if startedAt.IsZero() {
		return ""
	}
	if now.IsZero() {
		now = time.Now()
	}
	return humanElapsed(now.Sub(startedAt))
}

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

func humanDurationMs(durationMs int64) string {
	d := time.Duration(durationMs) * time.Millisecond
	if d < time.Second {
		return d.String()
	}
	return d.Round(time.Second).String()
}

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

func truncateHeader(s string) string {
	r := []rune(s)
	if len(r) <= headerMaxLen {
		return s
	}
	return string(r[:headerMaxLen-1]) + "…"
}

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

func nonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

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

func headerBlock(text string) *slack.HeaderBlock {
	return slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, truncateHeader(text), true, false))
}

func mrkdwnSection(text string) *slack.SectionBlock {
	return slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, text, false, false), nil, nil)
}

func fieldsSection(pairs []kv) *slack.SectionBlock {
	fields := make([]*slack.TextBlockObject, 0, len(pairs))
	for _, p := range pairs {
		fields = append(fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+p.k+"*\n"+p.v, false, false))
	}
	return slack.NewSectionBlock(nil, fields, nil)
}

func errorSection(e Event) (*slack.SectionBlock, bool) {
	if e.Outcome == nil || e.Outcome.Error == "" {
		return nil, false
	}
	return mrkdwnSection("*Error*\n```" + trimContext(e.Outcome.Error, 500) + "```"), true
}

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

func BuildComponentHealthMessage(e Event, recovered, installLevel bool) Message {
	headline := componentHealthHeadline(e, recovered, installLevel)

	blocks := []slack.Block{headerBlock(headline)}
	if fields := componentHealthFields(e, installLevel); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(componentHealthLinks(e, installLevel)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: headline, Blocks: blocks}
}

func componentHealthHeadline(e Event, recovered, installLevel bool) string {
	subject := metadataString(e, "component_name")
	if installLevel {
		subject = e.Workflow.OwnerName
	}

	var text string
	switch {
	case recovered && installLevel:
		text = "✅ Install recovered"
	case recovered:
		text = "✅ Component recovered"
	case installLevel:
		text = "🔴 Install degraded"
	case metadataString(e, "health") == "degraded":
		text = "⚠️ Component degraded"
	default:
		text = "🔴 Component unhealthy"
	}

	if !installLevel && !recovered && metadataString(e, "install_health") != "" {
		text += " · install degraded"
	}
	if !installLevel && recovered && metadataString(e, "install_health") != "" {
		text += " · install recovered"
	}

	if subject != "" {
		text = text + " · " + slackEscape(subject)
	}
	return text
}

func componentHealthFields(e Event, installLevel bool) []kv {
	fields := []kv{}

	if !installLevel {
		if name := metadataString(e, "component_name"); name != "" {
			fields = append(fields, kv{"Component", slackEscape(name)})
		}
	}
	if health := metadataString(e, "health"); health != "" {
		fields = append(fields, kv{"Health", slackEscape(health)})
	}
	if prev := metadataString(e, "previous_health"); prev != "" {
		fields = append(fields, kv{"Previously", slackEscape(prev)})
	}
	if msg := metadataString(e, "message"); msg != "" {
		fields = append(fields, kv{"Detail", slackEscape(msg)})
	}
	if resource := componentHealthResource(e); resource != "" {
		fields = append(fields, kv{"Resource", slackEscape(resource)})
	}
	if ih := metadataString(e, "install_health"); ih != "" && !installLevel {
		label := slackEscape(ih)
		if prev := metadataString(e, "install_previous_health"); prev != "" {
			label = slackEscape(prev) + " → " + label
		}
		fields = append(fields, kv{"Install health", label})
	}
	if name := e.Workflow.OwnerName; name != "" && !installLevel {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	return fields
}

func componentHealthResource(e Event) string {
	kind := metadataString(e, "root_resource_kind")
	name := metadataString(e, "root_resource_name")
	if kind == "" || name == "" {
		return ""
	}
	if ns := metadataString(e, "root_resource_namespace"); ns != "" {
		return kind + " " + ns + "/" + name
	}
	return kind + " " + name
}

func componentHealthLinks(e Event, installLevel bool) []LinkChip {
	if e.Links == nil {
		return nil
	}

	getters := []func(l *ContextLinks) string{
		func(l *ContextLinks) string { return l.Component },
		func(l *ContextLinks) string { return l.Install },
		func(l *ContextLinks) string { return l.Org },
	}
	if installLevel {
		getters = []func(l *ContextLinks) string{
			func(l *ContextLinks) string { return l.Install },
			func(l *ContextLinks) string { return l.Org },
		}
	}

	if l := firstNonEmptyLink(e.Links, getters...); l != "" {
		return []LinkChip{{Label: "Open in Nuon", URL: l}}
	}
	return nil
}

func BuildUpdateAppConfigMessage(e Event) Message {
	blocks := []slack.Block{headerBlock(updateAppConfigHeaderText(e))}
	if fields := updateAppConfigFields(e); len(fields) > 0 {
		blocks = append(blocks, slack.NewDividerBlock(), fieldsSection(fields))
	}
	if actions, ok := actionsBlock(updateAppConfigLinks(e)); ok {
		blocks = append(blocks, actions)
	}
	return Message{Text: plainUpdateAppConfigHeadline(e), Blocks: blocks}
}

func updateAppConfigHeaderText(e Event) string {
	text := "🔄 Install config updated"
	if name := metadataString(e, "install_name"); name != "" {
		text = text + " · " + name
	}
	return text
}

func updateAppConfigFields(e Event) []kv {
	var fields []kv
	if name := metadataString(e, "install_name"); name != "" {
		fields = append(fields, kv{"Install", slackEscape(name)})
	}
	if source := metadataString(e, "source"); source != "" {
		fields = append(fields, kv{"Source", slackEscape(source)})
	}
	if name := orgName(e); name != "" {
		fields = append(fields, kv{"Org", slackEscape(name)})
	}
	return fields
}

func updateAppConfigLinks(e Event) []LinkChip {
	if e.Links == nil {
		return nil
	}
	if l := firstNonEmptyLink(e.Links,
		func(l *ContextLinks) string { return l.Install },
		func(l *ContextLinks) string { return l.Org },
	); l != "" {
		return []LinkChip{{Label: "Open in Nuon", URL: l}}
	}
	return nil
}

func plainUpdateAppConfigHeadline(e Event) string {
	text := "🔄 Install config updated"
	if name := metadataString(e, "install_name"); name != "" {
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

func slackEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
