package awaitcomponenthealthy

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

type windowOutcome int

const (
	// windowWait means keep polling: nothing bad has been observed and the
	// window has not elapsed yet.
	windowWait windowOutcome = iota
	// windowPass means the window elapsed and every observation inside it was
	// healthy: the gate passes, at the window boundary, not a minute later.
	windowPass
	// windowFailBad means a bad observation landed inside the window — that's
	// fresh post-apply evidence, so fail now instead of waiting it out.
	windowFailBad
	// windowFailState means the window elapsed but the latest state (e.g.
	// progressing, unknown) never settled to healthy through the window.
	windowFailState
	// windowNeedData means the window elapsed with no observations at all —
	// the caller extends briefly rather than failing a good deploy on absent data.
	windowNeedData
)

func isBadReportHealth(health string) bool {
	return health == "degraded" || health == "unhealthy"
}

// judgeWindow judges raw reports (newest first) against the window, not the
// debounced verdict, so it concludes exactly when the window ends.
func judgeWindow(reports []activities.GateHealthReport, gateStartedAt, now time.Time, window time.Duration, requireCluster bool) (windowOutcome, *activities.GateHealthReport) {
	var inWindow []activities.GateHealthReport
	for _, r := range reports {
		if !time.Unix(r.ObservedAtTS, 0).Before(gateStartedAt) {
			inWindow = append(inWindow, r)
		}
	}

	for i := range inWindow {
		if isBadReportHealth(inWindow[i].Health) {
			return windowFailBad, &inWindow[i]
		}
	}

	// A passing probe proves an endpoint answered, not that the rollout
	// succeeded, so probe-only reports can't close the window on their own.
	if requireCluster {
		withCluster := make([]activities.GateHealthReport, 0, len(inWindow))
		for _, r := range inWindow {
			if r.ClusterEvidence {
				withCluster = append(withCluster, r)
			}
		}
		inWindow = withCluster
	}

	if now.Before(gateStartedAt.Add(window)) {
		if len(inWindow) > 0 {
			return windowWait, &inWindow[0]
		}
		return windowWait, nil
	}

	if len(inWindow) == 0 {
		return windowNeedData, nil
	}

	latest := &inWindow[0]
	if latest.Health == "healthy" {
		return windowPass, latest
	}
	return windowFailState, latest
}

// missingProbes returns declared probes absent from the window: the runner
// picks up new probes one cycle late, so a window closing there must not pass.
func missingProbes(declared []string, checks []activities.ComponentHealthCheckRow, gateStartedAt time.Time) []string {
	reported := map[string]bool{}
	for _, c := range checks {
		if c.ObservedAtTS <= 0 || time.Unix(c.ObservedAtTS, 0).Before(gateStartedAt) {
			continue
		}
		reported[c.Name] = true
	}

	var missing []string
	for _, name := range declared {
		if name != "" && !reported[name] {
			missing = append(missing, name)
		}
	}
	return missing
}

// withAwaitedChecks scopes the snapshot to in-window evidence: a verdict from
// before the apply says nothing about this deploy, so every check starts unknown
// and the first in-window report sets it.
func withAwaitedChecks(checks []activities.ComponentHealthCheckRow, declared []string, gateStartedAt time.Time) []activities.ComponentHealthCheckRow {
	present := map[string]bool{}
	out := make([]activities.ComponentHealthCheckRow, 0, len(checks)+len(declared))

	for _, c := range checks {
		present[c.Name] = true
		if c.ObservedAtTS <= 0 || time.Unix(c.ObservedAtTS, 0).Before(gateStartedAt) {
			c.Health = "unknown"
			c.Message = "no report since the deploy applied"
		}
		out = append(out, c)
	}

	for _, name := range declared {
		if present[name] {
			continue
		}
		out = append(out, activities.ComponentHealthCheckRow{
			Kind:    "CustomCheck",
			Name:    name,
			Health:  "unknown",
			Message: "waiting for a report",
		})
	}

	return out
}

// markRemovedCheckRows labels probe rows that are no longer declared in the
// component's current config, computed fresh each poll so a re-added probe
// loses the label immediately.
func markRemovedCheckRows(checks []activities.ComponentHealthCheckRow, declared []string) {
	set := map[string]bool{}
	for _, name := range declared {
		set[name] = true
	}
	for i := range checks {
		if isProbeKind(checks[i].Kind) && !set[checks[i].Name] {
			checks[i].Removed = true
		}
	}
}

func isProbeKind(kind string) bool {
	return strings.HasSuffix(kind, "Probe")
}

// failedChecks returns declared checks whose in-window report is failing.
// Presence alone is not enough: a required check exists so the deploy can be
// gated on its verdict, so one reporting degraded or unhealthy must fail the
// window rather than satisfy it.
func failedChecks(declared []string, checks []activities.ComponentHealthCheckRow, gateStartedAt time.Time) []string {
	want := map[string]bool{}
	for _, name := range declared {
		if name != "" {
			want[name] = true
		}
	}

	var failed []string
	for _, c := range checks {
		if !want[c.Name] {
			continue
		}
		if c.ObservedAtTS <= 0 || time.Unix(c.ObservedAtTS, 0).Before(gateStartedAt) {
			continue
		}
		if isBadReportHealth(c.Health) {
			failed = append(failed, c.Name+" is "+c.Health)
		}
	}
	sort.Strings(failed)
	return failed
}

// checkTransitions describes only what moved since the last poll, so a quiet
// window writes nothing to the timeline.
func checkTransitions(prev map[string]string, checks []activities.ComponentHealthCheckRow) string {
	var moved []string
	for _, c := range checks {
		was, seen := prev[c.Name]
		if !seen {
			// First sight of a check is only worth reporting once it says
			// something; unknown is the state everything starts in.
			if c.Health != "" && c.Health != "unknown" {
				moved = append(moved, fmt.Sprintf("%s unknown → %s", c.Name, c.Health))
			}
			continue
		}
		if was != c.Health {
			moved = append(moved, fmt.Sprintf("%s %s → %s", c.Name, was, c.Health))
		}
	}
	if len(moved) == 0 {
		return ""
	}
	sort.Strings(moved)
	return strings.Join(moved, ", ")
}

func rememberCheckHealth(prev map[string]string, checks []activities.ComponentHealthCheckRow) {
	for _, c := range checks {
		prev[c.Name] = c.Health
	}
}
