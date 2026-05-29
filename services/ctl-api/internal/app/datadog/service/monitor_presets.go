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
	preset app.DatadogManagedMonitorPreset,
	notifyHandles []string,
	displayName string,
) (ddclient.CreateMonitorRequest, error) {
	query, err := buildMonitorQuery(targetType, targetID, preset)
	if err != nil {
		return ddclient.CreateMonitorRequest{}, err
	}

	name := fmt.Sprintf("[Nuon] %s on %s %s", preset, targetType, targetIDLabel(targetID, displayName))

	message := buildMonitorMessage(targetType, targetID, preset, notifyHandles, displayName)

	return ddclient.CreateMonitorRequest{
		Name:    name,
		Type:    ddclient.MonitorTypeEventV2Alert,
		Query:   query,
		Message: message,
		Tags: []string{
			"source:nuon",
			"nuon_managed:true",
			fmt.Sprintf("nuon_target_type:%s", targetType),
			fmt.Sprintf("nuon_target_id:%s", targetID),
			fmt.Sprintf("nuon_preset:%s", preset),
		},
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
		// Action presets need a stable nuon_action_id tag that the
		// renderer doesn't emit yet — the workflow-run ID changes per
		// invocation, so reusing nuon_workflow_id here would only ever
		// match a single run. Tracked separately so we don't ship a
		// monitor that silently matches nothing.
		return "", fmt.Errorf("action target type is not yet supported (renderer enrichment pending)")
	default:
		return "", fmt.Errorf("unknown target type %q", targetType)
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
// boilerplate first, link to the Nuon resource second, @-handles last —
// matching DD's recommended message shape (DD only fans out @-mentions
// from the bottom of the message).
func buildMonitorMessage(
	targetType app.DatadogManagedMonitorTargetType,
	targetID string,
	preset app.DatadogManagedMonitorPreset,
	notifyHandles []string,
	displayName string,
) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Nuon-managed alert: %s on %s `%s`",
		preset, targetType, targetIDLabel(targetID, displayName))
	b.WriteString("\n\n")
	b.WriteString("Triggered by an event in the Nuon → Datadog event stream. ")
	b.WriteString("Created via the one-click \"Alert in Datadog\" action.")
	if len(notifyHandles) > 0 {
		b.WriteString("\n\n")
		b.WriteString(strings.Join(cleanHandles(notifyHandles), " "))
	}
	return b.String()
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
