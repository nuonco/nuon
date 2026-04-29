package generics

import "time"

func GetTimeDuration(startedAt time.Time, finishedAt time.Time) time.Duration {
	if startedAt.IsZero() {
		return time.Duration(0)
	} else if finishedAt.IsZero() {
		return time.Since(startedAt)
	}
	return finishedAt.Sub(startedAt)
}
