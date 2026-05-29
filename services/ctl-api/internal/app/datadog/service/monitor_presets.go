package service

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
)

// buildMonitorRequest assembles the DD monitor body for a given target +
// preset combo. The query uses DD's event-v2 search syntax against the
// stable nuon_* tags emitted by datadogrender.intrinsicTags — keeping
// the contract between the emitter and the monitor in one auditable
// place.
//
// Naming convention: monitor name is "[Nuon] <preset> on <target> <id>"
// so it sorts adjacent to other Nuon-managed monitors in the DD UI. The
// "[Nuon]" prefix is also a hint to ops that this monitor is API-managed
// and shouldn't be hand-edited (changes would be overwritten on the next
// sync — though v1 has no sync loop yet).
//
// Message splices the connection's DefaultNotifyHandles (or per-click
// override) into the alert body. DD's monitor model treats @-mentions in
// the message as routing directives, so this is how PagerDuty/Slack
// fan-out gets wired without us touching those integrations directly.
func buildMonitorRequest(
	targetType app.DatadogManagedMonitorTargetType,
	targetID string,
	installID string,
	preset app.DatadogManagedMonitorPreset,
	notifyHandles []string,
	displayName string,
	appURL string,
	orgID string,
) (ddclient.CreateMonitorRequest, error) {
	query, err := buildMonitorQuery(targetType, targetID, installID, preset)
	if err != nil {
		return ddclient.CreateMonitorRequest{}, err
	}

	name := fmt.Sprintf("[Nuon] %s on %s %s", preset, targetType, targetIDLabel(targetID, displayName))
	if installID != "" {
		name = fmt.Sprintf("%s (install %s)", name, installID)
	}

	message := buildMonitorMessage(targetType, targetID, installID, preset, notifyHandles, displayName, appURL, orgID)

	tags := []string{
		"source:nuon",
		"nuon_managed:true",
		fmt.Sprintf("nuon_target_type:%s", targetType),
		fmt.Sprintf("nuon_target_id:%s", targetID),
		fmt.Sprintf("nuon_preset:%s", preset),
	}
	if installID != "" {
		tags = append(tags, fmt.Sprintf("nuon_install_id:%s", installID))
	}

	return ddclient.CreateMonitorRequest{
		Name:    name,
		Type:    ddclient.MonitorTypeEventV2Alert,
		Query:   query,
		Message: message,
		Tags:    tags,
		Options: ddclient.MonitorOptions{
			NotifyNoData:     false,
			IncludeTags:      true,
			RenotifyInterval: 0,
		},
	}, nil
}

// buildMonitorQuery returns the DD event-v2 search string for the given
// target/preset. The leading "events(...)" wrapper plus the .rollup +
// threshold scaffolding is DD's standard shape for an event-stream
// monitor that fires on any matching event in the last 5 minutes.
//
// We deliberately key off the tag set that datadogrender.intrinsicTags
// always emits — adding a new target type means adding a stable nuon_*
// tag to the renderer first, otherwise the monitor would silently match
// nothing.
func buildMonitorQuery(
	targetType app.DatadogManagedMonitorTargetType,
	targetID string,
	installID string,
	preset app.DatadogManagedMonitorPreset,
) (string, error) {
	if targetID == "" {
		return "", fmt.Errorf("target_id is required")
	}

	var selector string
	switch targetType {
	case app.DatadogManagedMonitorTargetTypeInstall:
		selector = "nuon_install_id:" + targetID
	case app.DatadogManagedMonitorTargetTypeComponent:
		selector = "nuon_component_id:" + targetID
	case app.DatadogManagedMonitorTargetTypeWorkflow:
		selector = "nuon_workflow_id:" + targetID
	case app.DatadogManagedMonitorTargetTypeAction:
		// targetID is the org-level action_workflows.id, which is what
		// EventTargets.ActionID resolves to and what the DD renderer
		// stamps as nuon_action_id. The per-invocation
		// install_action_workflow_runs.id is intentionally not used
		// here — it'd only ever match a single run. Per-install
		// scoping rides on the installID parameter below, matching
		// nuon_install_id from the same renderer pass.
		selector = "nuon_action_id:" + targetID
		if installID == "" {
			// Require install scope on the dashboard's one-click path
			// for v1 — an org-wide action alert is a fine future
			// feature, but it should be an explicit opt-in rather
			// than something users land on by forgetting to pass
			// install_id.
			return "", fmt.Errorf("install_id is required for action target monitors")
		}
	default:
		return "", fmt.Errorf("unknown target type %q", targetType)
	}

	// Optional install scope ANDs nuon_install_id into the query so an
	// action monitor only fires for one install's invocations. Reuses
	// the install tag the renderer already stamps — no new tag needed.
	if installID != "" && targetType != app.DatadogManagedMonitorTargetTypeInstall {
		selector = selector + " nuon_install_id:" + installID
	}

	var conditionTag string
	switch preset {
	case app.DatadogManagedMonitorPresetFailure:
		// "failed" is the canonical nuon_status emitted by the
		// renderer's outcome mapping for failed phases. Keeping the
		// match on a stable enum value rather than the looser
		// transition tag (which carries cosmetic verbs like
		// "completed_with_errors") avoids false negatives when DD
		// strings drift.
		conditionTag = "nuon_status:failed"
	case app.DatadogManagedMonitorPresetDrift:
		// Drift events ride on signal_type:drift_detected — the
		// renderer maps it to nuon_kind:drift. The "failed" status
		// filter doesn't apply since drift is its own kind.
		conditionTag = "nuon_kind:drift"
	default:
		return "", fmt.Errorf("unknown preset %q", preset)
	}

	// events("source:nuon <selector> <condition>").rollup("count").last("5m") >= 1
	//
	// Why count >= 1: any single matching event should page. DD's
	// event-v2 monitor language treats this as "fire on first match,
	// recover after the 5m window clears" which matches user
	// expectations for a one-click failure alert.
	q := fmt.Sprintf(
		`events("source:nuon %s %s").rollup("count").last("5m") >= 1`,
		selector, conditionTag,
	)
	return q, nil
}

// targetIDLabel returns the display string used inside the monitor name.
// If the caller provided a display name (e.g. install slug or component
// name) we prefer it for readability; otherwise we fall back to the raw
// ID so the monitor name is always meaningful.
func targetIDLabel(targetID, displayName string) string {
	if strings.TrimSpace(displayName) != "" {
		return displayName
	}
	return targetID
}

// buildMonitorMessage produces the DD monitor body. Keeps Nuon
// boilerplate first, deep links to the Nuon resource(s) second,
// @-handles last — matching DD's recommended message shape (DD only
// fans out @-mentions from the bottom of the message).
//
// Deep links matter here because the alert lands in Slack / PagerDuty /
// email; the operator needs a one-click path back to the offending
// install or action. Without them, "view in Nuon" requires copy-pasting
// IDs from the alert into the dashboard URL bar.
func buildMonitorMessage(
	targetType app.DatadogManagedMonitorTargetType,
	targetID string,
	installID string,
	preset app.DatadogManagedMonitorPreset,
	notifyHandles []string,
	displayName string,
	appURL string,
	orgID string,
) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Nuon-managed alert: %s on %s `%s`",
		preset, targetType, targetIDLabel(targetID, displayName))
	b.WriteString("\n\n")
	b.WriteString("Triggered by an event in the Nuon → Datadog event stream. ")
	b.WriteString("Created via the one-click \"Alert in Datadog\" action.")

	if links := buildMonitorLinks(targetType, targetID, installID, appURL, orgID); len(links) > 0 {
		b.WriteString("\n\n")
		b.WriteString(strings.Join(links, "\n"))
	}

	if len(notifyHandles) > 0 {
		b.WriteString("\n\n")
		b.WriteString(strings.Join(cleanHandles(notifyHandles), " "))
	}
	return b.String()
}

// buildMonitorLinks returns the deep-link lines for the alert body.
//
// Most-specific link first (the offending resource), then the org
// dashboard as a fallback. Each link is a full markdown link — DD
// renders these clickably in Slack / PagerDuty / email when the monitor
// fires.
//
// Component and workflow targets without install scope can only build
// the org-level link. That's a real product gap (the dashboard groups
// those resources under installs), and a documented caveat of the
// monitor-create path: pass install_id whenever possible.
func buildMonitorLinks(
	targetType app.DatadogManagedMonitorTargetType,
	targetID string,
	installID string,
	appURL string,
	orgID string,
) []string {
	if appURL == "" || orgID == "" {
		return nil
	}
	base := strings.TrimRight(appURL, "/") + "/" + orgID

	var lines []string
	switch targetType {
	case app.DatadogManagedMonitorTargetTypeInstall:
		lines = append(lines, fmt.Sprintf("[Open install in Nuon](%s/installs/%s)", base, targetID))
	case app.DatadogManagedMonitorTargetTypeAction:
		if installID != "" {
			lines = append(lines, fmt.Sprintf("[Open action in Nuon](%s/installs/%s/actions/%s)", base, installID, targetID))
			lines = append(lines, fmt.Sprintf("[Open install in Nuon](%s/installs/%s)", base, installID))
		}
	case app.DatadogManagedMonitorTargetTypeComponent:
		if installID != "" {
			lines = append(lines, fmt.Sprintf("[Open component in Nuon](%s/installs/%s/components/%s)", base, installID, targetID))
			lines = append(lines, fmt.Sprintf("[Open install in Nuon](%s/installs/%s)", base, installID))
		}
	case app.DatadogManagedMonitorTargetTypeWorkflow:
		if installID != "" {
			lines = append(lines, fmt.Sprintf("[Open workflow in Nuon](%s/installs/%s/workflows/%s)", base, installID, targetID))
			lines = append(lines, fmt.Sprintf("[Open install in Nuon](%s/installs/%s)", base, installID))
		}
	}
	lines = append(lines, fmt.Sprintf("[Org dashboard](%s)", base))
	return lines
}

// cleanHandles trims whitespace and drops empty entries so a stray
// "" in notify_handles doesn't appear as "  " in the DD body.
func cleanHandles(in []string) []string {
	out := make([]string, 0, len(in))
	for _, h := range in {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		out = append(out, h)
	}
	return out
}
