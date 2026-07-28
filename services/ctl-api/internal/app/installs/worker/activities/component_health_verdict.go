package activities

import (
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	// componentHealthObservationWindow bounds how far back the evaluator reads
	// resource observations from ClickHouse. The runner reports every ~60s, so
	// this covers ~10 reports.
	componentHealthObservationWindow = 10 * time.Minute
	// componentHealthStaleAfter is ~3x the runner report interval: no report
	// inside it means the verdict is unknown, never unhealthy (absence of data
	// is not health data).
	componentHealthStaleAfter = 3 * time.Minute
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
}

// nextComponentHealthVerdict debounces a component's health verdict from its
// recent report history (newest first). A flip to degraded/unhealthy requires
// componentHealthFlipBadAfter consecutive bad reports; any other change
// requires componentHealthFlipAfter consecutive reports agreeing with the
// target; otherwise the current verdict holds. Components with no baseline
// verdict (unset, not-applicable, unknown) adopt the latest report
// immediately. No reports at all means not-applicable before a component has
// ever been observed, and unknown once it has (runner offline).
func nextComponentHealthVerdict(current app.InstallComponentHealthStatus, reports []componentHealthReport, now time.Time) app.InstallComponentHealthStatus {
	if len(reports) == 0 {
		if current == app.InstallComponentHealthStatusUnset || current == app.InstallComponentHealthStatusNotApplicable {
			return app.InstallComponentHealthStatusNotApplicable
		}
		return app.InstallComponentHealthStatusUnknown
	}

	latest := reports[0]
	if now.Sub(latest.ObservedAt) > componentHealthStaleAfter {
		return app.InstallComponentHealthStatusUnknown
	}

	target := latest.Health
	if target == current {
		return current
	}

	hasBaseline := current != app.InstallComponentHealthStatusUnset &&
		current != app.InstallComponentHealthStatusNotApplicable &&
		current != app.InstallComponentHealthStatusUnknown
	if !hasBaseline {
		return target
	}

	sevTarget := componentHealthSeverity[target]
	sevCurrent := componentHealthSeverity[current]
	sevDegraded := componentHealthSeverity[app.InstallComponentHealthStatusDegraded]

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

func componentHealthDescription(verdict app.InstallComponentHealthStatus, latest *componentHealthReport) string {
	switch verdict {
	case app.InstallComponentHealthStatusNotApplicable:
		return "component has no observable runtime resources"
	case app.InstallComponentHealthStatusUnknown:
		if latest == nil {
			return "no health observations reported"
		}
		return "no recent health observations from the runner"
	case app.InstallComponentHealthStatusHealthy:
		if latest != nil {
			if latest.Resources == 1 {
				return "1 resource healthy"
			}
			return fmt.Sprintf("all %d resources healthy", latest.Resources)
		}
		return "all resources healthy"
	default:
		if latest != nil && latest.RootKind != "" {
			root := latest.RootKind + " " + latest.RootName
			if latest.RootNamespace != "" {
				root = latest.RootKind + " " + latest.RootNamespace + "/" + latest.RootName
			}
			if latest.Message != "" {
				return root + ": " + latest.Message
			}
			return root + " is " + string(verdict)
		}
		return string(verdict)
	}
}
