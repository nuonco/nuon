// Package datadogrender renders Datadog event payloads for Nuon workflow /
// workflow_step / workflow_step_approval / drift-detected lifecycle events.
//
// Like slackrender, this is a pure function of its input event — no DB,
// no network. The signal lifecycle hook builds the input via webhook.go's
// enrichment pipeline (the same one slackrender consumes) and hands it to
// us along with the connection + subscription routing decisions already
// resolved.
//
// We deliberately reuse slackrender.Event as the input shape rather than
// declaring a parallel datadog-flavored type. The shape is identical, the
// underlying source (lifecycleEventData) is shared, and the hook already
// has a buildRenderEvent for it. The Slack-specific link fields
// (Approval, RespondAPI) are simply ignored on the DD side.
package datadogrender

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
)

// SourceName is the DD `source:` tag value Nuon always emits under. DD
// event-stream monitor queries filter on it (e.g.
// `events("source:nuon ...")`). Pinned here so the constant has one home.
const SourceName = "nuon"

// Build constructs a PostEventRequest for the given lifecycle event.
// Callers supply the routing-derived tag set (connection DefaultTags +
// subscription AdditionalTags + dispatch-time entity tags); Build merges
// them with the renderer's intrinsic tags (`source:nuon`, kind/transition,
// etc.) and dedupes.
//
// AggregationKey is set to the parent workflow ID so DD groups all step
// events under their workflow in the event stream UI — analog of Slack's
// thread anchor, but free.
//
// AlertTypeOverride / PriorityOverride win over the default mapping when
// non-empty. Validation of the override values is the caller's concern;
// DD itself rejects unknown values with a 400 which we surface as APIError.
func Build(e slackrender.Event, extraTags []string, alertTypeOverride, priorityOverride string) client.PostEventRequest {
	tags := mergeTags(intrinsicTags(e), extraTags)

	at := defaultAlertType(e)
	if alertTypeOverride != "" {
		at = client.EventAlertType(alertTypeOverride)
	}
	pr := defaultPriority(e)
	if priorityOverride != "" {
		pr = client.EventPriority(priorityOverride)
	}

	req := client.PostEventRequest{
		Title:          buildTitle(e),
		Text:           buildText(e),
		Tags:           tags,
		AlertType:      at,
		Priority:       pr,
		AggregationKey: aggregationKey(e),
		SourceTypeName: SourceName,
	}
	if !e.Workflow.CreatedAt.IsZero() {
		req.DateHappened = e.Workflow.CreatedAt.Unix()
	}
	return req
}

// intrinsicTags are the renderer-owned tags every event carries
// regardless of routing. Kept distinct from caller-supplied tags so the
// signal hook can't accidentally drop them.
//
// Tag keys mirror DD conventions: lowercase, colon-separated, no spaces.
func intrinsicTags(e slackrender.Event) []string {
	t := []string{
		"source:" + SourceName,
		"nuon_kind:" + e.Kind,
	}
	if e.Transition != "" {
		t = append(t, "nuon_transition:"+e.Transition)
	}
	if e.OrgID != "" {
		t = append(t, "nuon_org_id:"+e.OrgID)
	}
	if e.Workflow.ID != "" {
		t = append(t, "nuon_workflow_id:"+e.Workflow.ID)
	}
	if e.Workflow.Type != "" {
		t = append(t, "nuon_workflow_type:"+e.Workflow.Type)
	}
	if e.Workflow.OwnerType == slackrender.OwnerTypeInstalls && e.Workflow.OwnerID != "" {
		t = append(t, "nuon_install_id:"+e.Workflow.OwnerID)
	}
	if e.Step != nil {
		if e.Step.ID != "" {
			t = append(t, "nuon_step_id:"+e.Step.ID)
		}
		if e.Step.ComponentID != "" {
			t = append(t, "nuon_component_id:"+e.Step.ComponentID)
		}
	}
	if e.Outcome != nil && e.Outcome.Status != "" {
		t = append(t, "nuon_status:"+e.Outcome.Status)
	}
	return t
}

// mergeTags concatenates two tag slices, trims whitespace, drops empties,
// and dedupes preserving first-seen order. Sorted at the end so the
// payload is stable for snapshot tests / DD's own dedup heuristics.
func mergeTags(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, src := range [][]string{a, b} {
		for _, tag := range src {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, dup := seen[tag]; dup {
				continue
			}
			seen[tag] = struct{}{}
			out = append(out, tag)
		}
	}
	sort.Strings(out)
	return out
}

// defaultAlertType maps the lifecycle event onto a DD alert level. Picked
// to make the most common monitors "filter by alert_type" useful out of
// the box:
//
//   - explicit failure / cancelled → error / warning
//   - approval request → warning ("attention needed")
//   - drift detected → warning ("attention needed")
//   - everything else → info
func defaultAlertType(e slackrender.Event) client.EventAlertType {
	if e.Outcome != nil {
		switch e.Outcome.Status {
		case slackrender.TransitionFailed:
			return client.EventAlertTypeError
		case slackrender.TransitionCancelled:
			return client.EventAlertTypeWarning
		case slackrender.TransitionSucceeded:
			return client.EventAlertTypeSuccess
		}
	}
	if e.Kind == slackrender.KindWorkflowStepApproval && e.Transition == slackrender.TransitionRequested {
		return client.EventAlertTypeWarning
	}
	// Drift-detected events have Kind="workflow" with Type="drift_run"
	// and an explicit transition; the workflow_step that carries the
	// drift signal is currently encoded with a distinct status the
	// signal hook surfaces via Outcome. Fallback to info covers the
	// "informational lifecycle update" case.
	return client.EventAlertTypeInfo
}

// defaultPriority leaves DD's "normal" priority for everything except
// non-actionable "started" / "succeeded" events, which we mark "low" so
// they don't compete for monitor attention with real incidents.
func defaultPriority(e slackrender.Event) client.EventPriority {
	switch e.Transition {
	case slackrender.TransitionStarted, slackrender.TransitionSucceeded:
		return client.EventPriorityLow
	}
	if e.Outcome != nil && e.Outcome.Status == slackrender.TransitionSucceeded {
		return client.EventPriorityLow
	}
	return client.EventPriorityNormal
}

// aggregationKey is the workflow ID — DD's event stream UI groups all
// events sharing an aggregation_key into a single rolled-up entry, which
// is the equivalent of Slack's threaded posts for our taxonomy. Nested
// action_workflow_run sub-workflows roll up under their launching deploy
// step's workflow via ParentRef so the user-visible run stays singular.
func aggregationKey(e slackrender.Event) string {
	if e.Parent != nil && e.Parent.WorkflowID != "" {
		return e.Parent.WorkflowID
	}
	return e.Workflow.ID
}

// buildTitle constructs the DD event title (the bold one-liner shown in
// the stream). Format mirrors what users will read in a monitor's
// notification message:
//
//	"Nuon: provision install <name> failed"
//	"Nuon: terraform-plan step on install <name> requires approval"
//	"Nuon: drift detected on install <name>"
//
// The Nuon prefix is structural — it makes browsing the DD event stream
// scannable when mixed with non-Nuon events.
func buildTitle(e slackrender.Event) string {
	var parts []string
	parts = append(parts, "Nuon:")

	wfType := e.Workflow.Type
	if wfType == "" {
		wfType = "workflow"
	}

	switch e.Kind {
	case slackrender.KindWorkflowStep:
		stepName := "step"
		if e.Step != nil && e.Step.Name != "" {
			stepName = e.Step.Name
		}
		parts = append(parts, stepName, "step")
	case slackrender.KindWorkflowStepApproval:
		stepName := "step"
		if e.Step != nil && e.Step.Name != "" {
			stepName = e.Step.Name
		}
		parts = append(parts, "approval", "for", stepName)
	default:
		parts = append(parts, wfType)
	}

	if e.Workflow.OwnerName != "" {
		ownerKind := strings.TrimSuffix(e.Workflow.OwnerType, "s")
		if ownerKind == "" {
			ownerKind = "owner"
		}
		parts = append(parts, "on", ownerKind, e.Workflow.OwnerName)
	}

	verb := transitionVerb(e)
	if verb != "" {
		parts = append(parts, verb)
	}
	return strings.Join(parts, " ")
}

// transitionVerb returns the past-tense verb for the title. The string
// here is what shows up after the entity reference (e.g. "failed").
func transitionVerb(e slackrender.Event) string {
	if e.Outcome != nil {
		switch e.Outcome.Status {
		case slackrender.TransitionFailed:
			return "failed"
		case slackrender.TransitionSucceeded:
			return "succeeded"
		case slackrender.TransitionCancelled:
			return "was cancelled"
		}
	}
	switch e.Transition {
	case slackrender.TransitionStarted:
		return "started"
	case slackrender.TransitionRequested:
		return "is awaiting approval"
	case slackrender.TransitionApproved:
		return "was approved"
	case slackrender.TransitionRejected:
		return "was rejected"
	}
	return ""
}

// buildText constructs the DD event body in Markdown — DD supports a
// limited Markdown subset (`%%%` … `%%%` for the block) and renders
// links / bold inline. We surface the operator-relevant bits without
// duplicating the title:
//
//	**Workflow:** <id>
//	**Started by:** alice@example.com
//	**Error:** <truncated>
//	**Open in Nuon:** <link>
//	**Open in DD:** computed by the caller if it wants to deep-link
func buildText(e slackrender.Event) string {
	var b strings.Builder
	b.WriteString("%%%\n")

	if e.Workflow.ID != "" {
		fmt.Fprintf(&b, "**Workflow:** `%s`\n", e.Workflow.ID)
	}
	if e.Workflow.Type != "" {
		fmt.Fprintf(&b, "**Type:** %s\n", e.Workflow.Type)
	}
	if e.Workflow.RunbookName != "" {
		fmt.Fprintf(&b, "**Runbook:** %s\n", e.Workflow.RunbookName)
	}
	if e.Step != nil && e.Step.Name != "" {
		fmt.Fprintf(&b, "**Step:** %s\n", e.Step.Name)
	}
	if e.Step != nil && e.Step.ComponentName != "" {
		fmt.Fprintf(&b, "**Component:** %s\n", e.Step.ComponentName)
	}
	if e.Workflow.CreatedByEmail != "" {
		fmt.Fprintf(&b, "**Started by:** %s\n", e.Workflow.CreatedByEmail)
	}
	if !e.Workflow.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "**Started:** %s\n", e.Workflow.CreatedAt.UTC().Format(time.RFC3339))
	}
	if e.Outcome != nil {
		if e.Outcome.DurationMs > 0 {
			fmt.Fprintf(&b, "**Duration:** %s\n", time.Duration(e.Outcome.DurationMs)*time.Millisecond)
		}
		if e.Outcome.Error != "" {
			fmt.Fprintf(&b, "**Error:** %s\n", truncate(e.Outcome.Error, 1024))
		}
	}
	if e.Approval != nil {
		if e.Approval.Type != "" {
			fmt.Fprintf(&b, "**Approval type:** %s\n", e.Approval.Type)
		}
		if e.Approval.RespondedBy != "" {
			fmt.Fprintf(&b, "**Responded by:** %s\n", e.Approval.RespondedBy)
		}
		if e.Approval.Plan != "" {
			fmt.Fprintf(&b, "\n```\n%s\n```\n", truncate(e.Approval.Plan, 2048))
		}
	}
	if e.Links != nil && e.Links.Workflow != "" {
		fmt.Fprintf(&b, "\n[Open in Nuon](%s)\n", e.Links.Workflow)
	}
	b.WriteString("%%%")
	return b.String()
}

// truncate shortens a string to at most n bytes, suffixing "..." when
// truncation occurred. Used for error / plan excerpts to keep DD payloads
// well under their 4000-byte text limit.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
