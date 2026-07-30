package service

import (
	"math"
	"sort"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	healthTimelineDefaultDays = 90
	healthTimelineMaxDays     = 90
	healthTimelineMinDays     = 1
)

var healthUnknown = string(app.InstallComponentHealthStatusUnknown)

func clampHealthTimelineDays(days int) int {
	switch {
	case days < healthTimelineMinDays:
		return healthTimelineMinDays
	case days > healthTimelineMaxDays:
		return healthTimelineMaxDays
	default:
		return days
	}
}

// healthWindow returns the [from, to) window for a days-sized timeline
// ending now, with `from` at UTC midnight so buckets align to calendar dates.
func healthWindow(now time.Time, days int) (from, to time.Time) {
	now = now.UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return todayStart.AddDate(0, 0, -(days - 1)), now
}

type healthSpan struct {
	From   time.Time
	To     time.Time
	Health string
}

// healthSpans materializes the verdict in effect at every instant across
// [from, to) from a sparse list of transition edges. initialHealth seeds the
// window so it doesn't read as unknown before the first transition inside it.
func healthSpans(transitions []app.InstallComponentHealthTransition, from, to time.Time, initialHealth string) []healthSpan {
	if !to.After(from) {
		return nil
	}
	if initialHealth == "" {
		initialHealth = healthUnknown
	}

	sorted := make([]app.InstallComponentHealthTransition, 0, len(transitions))
	for _, t := range transitions {
		if t.ObservedAt.Before(from) || !t.ObservedAt.Before(to) {
			continue
		}
		sorted = append(sorted, t)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ObservedAt.Before(sorted[j].ObservedAt)
	})

	spans := make([]healthSpan, 0, len(sorted)+1)
	cursor := from
	currentHealth := initialHealth

	for _, t := range sorted {
		if t.ObservedAt.After(cursor) {
			spans = append(spans, healthSpan{From: cursor, To: t.ObservedAt, Health: currentHealth})
		}
		cursor = t.ObservedAt
		currentHealth = t.ToHealth
	}
	if to.After(cursor) {
		spans = append(spans, healthSpan{From: cursor, To: to, Health: currentHealth})
	}

	return spans
}

// dailySeverity ranks verdicts for picking the worst one in a window:
// unhealthy > degraded > unknown > progressing > healthy. Not-applicable and
// unset map to unknown's rank here.
func dailySeverity(health string) int {
	switch health {
	case string(app.InstallComponentHealthStatusUnhealthy):
		return 4
	case string(app.InstallComponentHealthStatusDegraded):
		return 3
	case string(app.InstallComponentHealthStatusProgressing):
		return 1
	case string(app.InstallComponentHealthStatusHealthy):
		return 0
	default:
		return 2
	}
}

func normalizeHealthLabel(health string) string {
	switch health {
	case string(app.InstallComponentHealthStatusUnhealthy),
		string(app.InstallComponentHealthStatusDegraded),
		string(app.InstallComponentHealthStatusProgressing),
		string(app.InstallComponentHealthStatusHealthy):
		return health
	default:
		return healthUnknown
	}
}

// healthTotals accumulates span seconds by verdict. Not-applicable spans
// fold into UnknownSeconds since both mean "no signal" for uptime purposes.
type healthTotals struct {
	HealthySeconds     int64
	ProgressingSeconds int64
	DegradedSeconds    int64
	UnhealthySeconds   int64
	UnknownSeconds     int64
}

func (t healthTotals) observedSeconds() int64 {
	return t.HealthySeconds + t.ProgressingSeconds + t.DegradedSeconds + t.UnhealthySeconds
}

// uptimePercent is the fraction of observed time (excluding unknown) that
// wasn't degraded/unhealthy. Zero observed time reports 0 rather than
// dividing by zero — pair with observedSeconds to tell "no data" from "0% up".
func (t healthTotals) uptimePercent() float64 {
	observed := t.observedSeconds()
	if observed == 0 {
		return 0
	}
	bad := t.DegradedSeconds + t.UnhealthySeconds
	return roundPercent(float64(observed-bad) / float64(observed) * 100)
}

func roundPercent(pct float64) float64 {
	return math.Round(pct*100) / 100
}

func addSpanSeconds(totals *healthTotals, health string, seconds int64) {
	switch health {
	case string(app.InstallComponentHealthStatusHealthy):
		totals.HealthySeconds += seconds
	case string(app.InstallComponentHealthStatusProgressing):
		totals.ProgressingSeconds += seconds
	case string(app.InstallComponentHealthStatusDegraded):
		totals.DegradedSeconds += seconds
	case string(app.InstallComponentHealthStatusUnhealthy):
		totals.UnhealthySeconds += seconds
	default:
		totals.UnknownSeconds += seconds
	}
}

func sumSpans(spans []healthSpan) healthTotals {
	var totals healthTotals
	for _, s := range spans {
		addSpanSeconds(&totals, s.Health, int64(s.To.Sub(s.From).Seconds()))
	}
	return totals
}

type dailyHealthBucket struct {
	Date             string `json:"date"`
	Health           string `json:"health"`
	UnhealthySeconds int64  `json:"unhealthy_seconds"`
	DegradedSeconds  int64  `json:"degraded_seconds"`
	UnknownSeconds   int64  `json:"unknown_seconds"`
	ObservedSeconds  int64  `json:"observed_seconds"`
}

// foldDailyHealth buckets spans into exactly `days` calendar-day rows from
// `from`, splitting spans that cross a day boundary. A day's health is the
// worst verdict seen; a day with no data at all (vs. an unknown verdict) is left "".
func foldDailyHealth(spans []healthSpan, from time.Time, days int) []dailyHealthBucket {
	buckets := make([]dailyHealthBucket, days)
	totals := make([]healthTotals, days)
	worstSeverity := make([]int, days)
	worstHealth := make([]string, days)
	hasData := make([]bool, days)

	for i := 0; i < days; i++ {
		buckets[i].Date = from.AddDate(0, 0, i).Format("2006-01-02")
		worstSeverity[i] = -1
	}

	for _, span := range spans {
		// Not-applicable/unset spans carry no signal and must not outvote a healthy day.
		if !spanBearsSignal(span.Health) {
			continue
		}
		cursor := span.From
		for cursor.Before(span.To) {
			dayIdx := int(cursor.Sub(from) / (24 * time.Hour))
			if dayIdx < 0 {
				dayIdx = 0
			}
			if dayIdx >= days {
				break
			}

			dayEnd := from.AddDate(0, 0, dayIdx+1)
			segEnd := span.To
			if segEnd.After(dayEnd) {
				segEnd = dayEnd
			}

			seconds := int64(segEnd.Sub(cursor).Seconds())
			if seconds > 0 {
				addSpanSeconds(&totals[dayIdx], span.Health, seconds)
				hasData[dayIdx] = true
				if sev := dailySeverity(span.Health); sev > worstSeverity[dayIdx] {
					worstSeverity[dayIdx] = sev
					worstHealth[dayIdx] = normalizeHealthLabel(span.Health)
				}
			}
			cursor = segEnd
		}
	}

	for i := 0; i < days; i++ {
		buckets[i].UnhealthySeconds = totals[i].UnhealthySeconds
		buckets[i].DegradedSeconds = totals[i].DegradedSeconds
		buckets[i].UnknownSeconds = totals[i].UnknownSeconds
		buckets[i].ObservedSeconds = totals[i].observedSeconds()
		if hasData[i] {
			buckets[i].Health = worstHealth[i]
		}
	}

	return buckets
}

// worstDailyAcrossComponents rolls per-component daily buckets into one row
// per day: whichever component had the worst verdict that day. No-data days stay "".
func worstDailyAcrossComponents(perComponent [][]dailyHealthBucket, from time.Time, days int) []dailyHealthBucket {
	out := make([]dailyHealthBucket, days)
	for i := 0; i < days; i++ {
		out[i].Date = from.AddDate(0, 0, i).Format("2006-01-02")

		bestSeverity := -1
		for _, buckets := range perComponent {
			if i >= len(buckets) || buckets[i].Health == "" {
				continue
			}
			if sev := dailySeverity(buckets[i].Health); sev > bestSeverity {
				bestSeverity = sev
				out[i] = buckets[i]
			}
		}
	}
	return out
}

func spanBearsSignal(health string) bool {
	switch health {
	case string(app.InstallComponentHealthStatusUnset),
		string(app.InstallComponentHealthStatusNotApplicable):
		return false
	}
	return true
}
