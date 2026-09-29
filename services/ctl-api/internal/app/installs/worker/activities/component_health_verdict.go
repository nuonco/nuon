package activities

import (
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	componentHealthObservationWindow = 10 * time.Minute
	componentHealthProgressingLimit  = 30 * time.Minute
	componentHealthStaleAfter        = 5 * time.Minute

	customCheckRetentionWindow  = 60 * time.Minute
	componentHealthFlipBadAfter = 3
	componentHealthFlipAfter    = 2
)

var componentHealthSeverity = map[app.InstallComponentHealthStatus]int{
	app.InstallComponentHealthStatusHealthy:     0,
	app.InstallComponentHealthStatusUnknown:     1,
	app.InstallComponentHealthStatusProgressing: 1,
	app.InstallComponentHealthStatusDegraded:    2,
	app.InstallComponentHealthStatusUnhealthy:   3,
}

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

	ClusterEvidence bool

	ValidFor time.Duration
}

func (r componentHealthReport) validFor() time.Duration {
	if r.ValidFor <= 0 {
		return componentHealthStaleAfter
	}
	return r.ValidFor
}

func nextComponentHealthVerdict(current app.InstallComponentHealthStatus, reports []componentHealthReport, now time.Time) app.InstallComponentHealthStatus {
	if len(reports) == 0 {
		if current == app.InstallComponentHealthStatusUnset || current == app.InstallComponentHealthStatusNotApplicable {
			return app.InstallComponentHealthStatusNotApplicable
		}
		return app.InstallComponentHealthStatusUnknown
	}

	latest := reports[0]
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

func rootKindRank(kind string) int {
	switch kind {
	case "Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob":
		return 0
	case "ReplicaSet":
		return 2
	case "Pod":
		return 3
	}
	return 1
}

func betterRoot(cur *componentHealthReport, health app.InstallComponentHealthStatus, kind, namespace, name string) bool {
	if s, c := componentHealthSeverity[health], componentHealthSeverity[cur.Health]; s != c {
		return s > c
	}
	if r, c := rootKindRank(kind), rootKindRank(cur.RootKind); r != c {
		return r < c
	}
	if kind != cur.RootKind {
		return kind < cur.RootKind
	}
	if namespace != cur.RootNamespace {
		return namespace < cur.RootNamespace
	}
	return name < cur.RootName
}

func otherAffected(rep *componentHealthReport, health app.InstallComponentHealthStatus) int {
	sev := componentHealthSeverity[health]
	if sev < componentHealthSeverity[app.InstallComponentHealthStatusDegraded] {
		return 0
	}
	n := 0
	for h, c := range rep.ResourceCounts {
		if componentHealthSeverity[app.InstallComponentHealthStatus(h)] >= sev {
			n += c
		}
	}
	return n - 1
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
		return rootResourceDescription(latest, app.InstallComponentHealthStatusUnknown)
	case app.InstallComponentHealthStatusHealthy:
		if latest != nil && latest.Health != app.InstallComponentHealthStatusHealthy {
			return rootResourceDescription(latest, latest.Health) + " — confirming before the status changes"
		}
		if latest != nil {
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

func rootResourceDescription(latest *componentHealthReport, health app.InstallComponentHealthStatus) string {
	if latest == nil || latest.RootKind == "" {
		return string(health)
	}

	root := latest.RootKind + " " + latest.RootName
	if latest.RootNamespace != "" {
		root = latest.RootKind + " " + latest.RootNamespace + "/" + latest.RootName
	}
	desc := root + " is " + string(health)
	if latest.Message != "" {
		desc = root + ": " + latest.Message
	}
	if others := otherAffected(latest, health); others > 0 {
		desc += fmt.Sprintf(" (+%d more affected)", others)
	}
	return desc
}
