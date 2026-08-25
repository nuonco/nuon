package activities

import (
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	// componentHealthObservationWindow bounds how far back the evaluator reads
	// observations from ClickHouse (~10 reports at the runner's ~60s cadence).
	componentHealthObservationWindow = 10 * time.Minute
	// componentHealthProgressingLimit is how long a component may stay
	// progressing before it is treated as degraded. Long enough for a slow
	// rollout or a cluster autoscaler cold start, short enough that a genuinely
	// stuck workload is reported the same day.
	componentHealthProgressingLimit = 30 * time.Minute
	// componentHealthStaleAfter marks a stale verdict unknown, never unhealthy —
	// absence of data isn't health data. Matches the 5m runner-inactive cutoff.
	componentHealthStaleAfter = 5 * time.Minute

	// customCheckRetentionWindow is the longest TTL a pushed check may declare,
	// and so how far back they are read.
	customCheckRetentionWindow = 60 * time.Minute
	// componentHealthFlipBadAfter is how many consecutive bad (degraded or
	// worse) reports it takes to flip a component's verdict bad.
	componentHealthFlipBadAfter = 3
	// componentHealthFlipAfter is how many consecutive agreeing reports any
	// other verdict change takes (e.g. recovery to healthy).
	componentHealthFlipAfter = 2
)

// componentHealthSeverity mirrors the ranking the ingest latch uses.
var componentHealthSeverity = map[app.InstallComponentHealthStatus]int{
	app.InstallComponentHealthStatusHealthy:     0,
	app.InstallComponentHealthStatusUnknown:     1,
	app.InstallComponentHealthStatusProgressing: 1,
	app.InstallComponentHealthStatusDegraded:    2,
	app.InstallComponentHealthStatusUnhealthy:   3,
}

// componentHealthReport is one runner report for a component: the worst
// resource across everything observed at a single report timestamp.
type componentHealthReport struct {
	ObservedAt time.Time
	Health     app.InstallComponentHealthStatus

	RootKind      string
	RootNamespace string
	RootName      string
	Message       string
	NativeStatus  string

	Resources      int
	ResourceCounts map[string]int

	// ClusterEvidence is true if this report has any cluster-derived
	// observation; distinguishes "looks fine" from "can't see the workload".
	ClusterEvidence bool

	// ValidFor is how long this report stays trustworthy; zero means the
	// default. A report synthesized from a pushed check inherits that check's
	// window, since the runner's clock says nothing about it.
	ValidFor time.Duration
}

func (r componentHealthReport) validFor() time.Duration {
	if r.ValidFor <= 0 {
		return componentHealthStaleAfter
	}
	return r.ValidFor
}

// nextComponentHealthVerdict debounces a verdict against its recent report
// history; components with no baseline verdict adopt the latest report
// immediately. No reports means not-applicable if never observed, else unknown.
func nextComponentHealthVerdict(current app.InstallComponentHealthStatus, reports []componentHealthReport, now time.Time) app.InstallComponentHealthStatus {
	if len(reports) == 0 {
		if current == app.InstallComponentHealthStatusUnset || current == app.InstallComponentHealthStatusNotApplicable {
			return app.InstallComponentHealthStatusNotApplicable
		}
		return app.InstallComponentHealthStatusUnknown
	}

	latest := reports[0]
	// Measured against the report's own window, not a global constant: a
	// pushed check declaring 30m must not be discarded at 5m.
	if now.Sub(latest.ObservedAt) > latest.validFor() {
		return app.InstallComponentHealthStatusUnknown
	}

	target := latest.Health
	if target == current {
		return current
	}

	sevTarget := componentHealthSeverity[target]
	sevCurrent := componentHealthSeverity[current]
	sevDegraded := componentHealthSeverity[app.InstallComponentHealthStatusDegraded]

	// Fast to good, slow to bad: a component fresh from a deploy has no baseline,
	// and claiming bad on its first report made every transient an outage.
	hasBaseline := current != app.InstallComponentHealthStatusUnset &&
		current != app.InstallComponentHealthStatusNotApplicable &&
		current != app.InstallComponentHealthStatusUnknown
	if !hasBaseline && sevTarget < sevDegraded {
		return target
	}

	required := componentHealthFlipAfter
	agrees := func(r componentHealthReport) bool {
		return componentHealthSeverity[r.Health] <= sevTarget
	}
	if sevTarget > sevCurrent {
		threshold := sevTarget
		if sevTarget >= sevDegraded {
			required = componentHealthFlipBadAfter
			threshold = sevDegraded
		}
		agrees = func(r componentHealthReport) bool {
			return componentHealthSeverity[r.Health] >= threshold
		}
	}

	if len(reports) < required {
		return current
	}
	for _, r := range reports[:required] {
		if !agrees(r) {
			return current
		}
	}

	return target
}

func componentHealthDescription(verdict app.InstallComponentHealthStatus, latest *componentHealthReport, now time.Time) string {
	switch verdict {
	case app.InstallComponentHealthStatusNotApplicable:
		return "component has no observable runtime resources"
	case app.InstallComponentHealthStatusUnknown:
		if latest == nil {
			return "no health observations reported"
		}
		if now.Sub(latest.ObservedAt) > latest.validFor() {
			return "no recent health observations from the runner"
		}
		// Observations arrived but none could be assessed — name what couldn't
		// be checked rather than blaming the runner.
		return rootResourceDescription(latest, app.InstallComponentHealthStatusUnknown)
	case app.InstallComponentHealthStatusHealthy:
		// The debounce can hold healthy against a newer worse observation —
		// describe what was actually seen instead of contradicting it.
		if latest != nil && latest.Health != app.InstallComponentHealthStatusHealthy {
			return rootResourceDescription(latest, latest.Health) + " — confirming before the status changes"
		}
		if latest != nil {
			// Resources includes unchecked ones; folding them into "all healthy"
			// would call a check that never ran passing — report separately.
			unchecked := latest.ResourceCounts[string(app.InstallComponentHealthStatusUnknown)]
			if unchecked > 0 {
				return fmt.Sprintf("%d of %d resources healthy, %d could not be checked",
					latest.Resources-unchecked, latest.Resources, unchecked)
			}
			if latest.Resources == 1 {
				return "1 resource healthy"
			}
			return fmt.Sprintf("all %d resources healthy", latest.Resources)
		}
		return "all resources healthy"
	default:
		return rootResourceDescription(latest, verdict)
	}
}

// rootResourceDescription names the resource responsible for a verdict and why.
func rootResourceDescription(latest *componentHealthReport, health app.InstallComponentHealthStatus) string {
	if latest == nil || latest.RootKind == "" {
		return string(health)
	}

	root := latest.RootKind + " " + latest.RootName
	if latest.RootNamespace != "" {
		root = latest.RootKind + " " + latest.RootNamespace + "/" + latest.RootName
	}
	if latest.Message != "" {
		return root + ": " + latest.Message
	}
	return root + " is " + string(health)
}
