package common

import (
	"fmt"
	"math"
)

// Handles simple humanization of durations in the orders of seconds to hours
func HumanizeNSDuration(nanoseconds int64) string {
	if nanoseconds == 0 {
		return "0s"
	}
	dur := float64(nanoseconds) / 1e9
	if dur < 60 {
		return fmt.Sprintf("%ds", int(dur))
	}
	durMin := math.Floor(dur / 60)
	durMinSec := math.Mod(dur, 60)
	if durMin < 60 {
		return fmt.Sprintf("%dm %ds", int64(durMin), int64(durMinSec))
	}
	durHour := int64(math.Floor(durMin / 60))
	durHourMin := int64(math.Mod(durMin, 60))
	return fmt.Sprintf("%dm %ds", durHour, durHourMin)
}
