package generics

import "time"

func GetTimeDuration(startedAt time.Time, finishedAt time.Time) time.Duration {
	if finishedAt.IsZero() && startedAt.IsZero() {
		return time.Duration(0)
	} else if !startedAt.IsZero() && finishedAt.IsZero() {
		return time.Now().Sub(startedAt)
	}
	return finishedAt.Sub(startedAt)
}
