package cronutil

/**
This is a temporary package made for adding jitter to cron schedules, ideally we wanna move to use temporal cron
schedule api which has built in jitter support. This is currently a bit hacky to parse cron schedule but works for most
of the things we have in the system.
**/

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// MaxJitterWindow requests the widest spread ApplyCronJitter can express;
// the effective jitter is still capped at each schedule's firing interval.
const MaxJitterWindow = time.Hour

const minIntervalProbeFires = 200

// ApplyCronJitter deterministically shifts the minute field of a standard
// 5-field cron expression by hash(emitterID) minutes. Mirrors Temporal
// schedule jitter semantics: window is a maximum — the effective jitter is
// capped at the schedule's shortest firing interval and at one hour. Never
// fails; anything unsupported returns the schedule unchanged.
func ApplyCronJitter(emitterID, schedule string, window time.Duration) string {
	if window <= 0 || emitterID == "" {
		return schedule
	}

	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return schedule
	}
	if interval := MinScheduleInterval(sched); interval < window {
		window = interval
	}
	if window > MaxJitterWindow {
		window = MaxJitterWindow
	}
	windowMinutes := int(window / time.Minute)
	if windowMinutes <= 0 {
		return schedule
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(emitterID))
	offset := int(h.Sum32() % uint32(windowMinutes))
	if offset == 0 {
		return schedule
	}

	fields := strings.Fields(schedule)
	if len(fields) != 5 {
		return schedule
	}

	minute, ok := shiftMinuteField(fields[0], offset)
	if !ok || minute == fields[0] {
		return schedule
	}

	fields[0] = minute
	jittered := strings.Join(fields, " ")
	if _, err := cron.ParseStandard(jittered); err != nil {
		return schedule
	}
	return jittered
}

// MinScheduleInterval returns the smallest gap between two consecutive fires
// of the given schedule, sampled over the next minIntervalProbeFires fires.
func MinScheduleInterval(sched cron.Schedule) time.Duration {
	now := time.Now().UTC()
	prev := sched.Next(now)
	if prev.IsZero() {
		return time.Duration(1<<63 - 1)
	}
	min := time.Duration(1<<63 - 1)
	for i := 0; i < minIntervalProbeFires; i++ {
		next := sched.Next(prev)
		if next.IsZero() {
			break
		}
		if d := next.Sub(prev); d < min {
			min = d
		}
		prev = next
	}
	return min
}

func shiftMinuteField(field string, offset int) (string, bool) {
	switch {
	case field == "*":
		return field, true
	case strings.HasPrefix(field, "*/"):
		step, err := strconv.Atoi(field[2:])
		if err != nil || step < 1 || step > 59 {
			return "", false
		}
		return fmt.Sprintf("%d/%d", offset, step), true
	case strings.Contains(field, ","):
		parts := strings.Split(field, ",")
		shifted := make([]int, 0, len(parts))
		seen := make(map[int]bool, len(parts))
		for _, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n > 59 {
				return "", false
			}
			v := (n + offset) % 60
			if !seen[v] {
				seen[v] = true
				shifted = append(shifted, v)
			}
		}
		sort.Ints(shifted)
		out := make([]string, len(shifted))
		for i, v := range shifted {
			out[i] = strconv.Itoa(v)
		}
		return strings.Join(out, ","), true
	default:
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 || n > 59 {
			return "", false
		}
		return strconv.Itoa((n + offset) % 60), true
	}
}
