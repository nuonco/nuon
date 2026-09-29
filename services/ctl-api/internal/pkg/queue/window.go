package queue

import (
	"strings"
	"time"
)

type ReleaseWindow struct {
	Days      []string
	StartTime string
	EndTime   string
	Timezone  string
}

func (w *ReleaseWindow) IsOpen(t time.Time) bool {
	loc, err := time.LoadLocation(w.Timezone)
	if err != nil {
		loc = time.UTC
	}
	t = t.In(loc)

	dayMatch := false
	currentDay := t.Weekday().String()
	for _, day := range w.Days {
		if strings.EqualFold(day[:3], currentDay[:3]) {
			dayMatch = true
			break
		}
	}
	if !dayMatch {
		return false
	}

	start, err := time.Parse("15:04", w.StartTime)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", w.EndTime)
	if err != nil {
		return false
	}

	startTime := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), 0, 0, loc)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), end.Hour(), end.Minute(), 0, 0, loc)

	return (t.Equal(startTime) || t.After(startTime)) && t.Before(endTime)
}

func (w *ReleaseWindow) NextOpenTime(t time.Time) time.Time {
	if w.IsOpen(t) {
		return t
	}

	loc, err := time.LoadLocation(w.Timezone)
	if err != nil {
		loc = time.UTC
	}
	t = t.In(loc)

	for i := 0; i < 8; i++ {
		dayMatch := false
		currentDay := t.Weekday().String()
		for _, day := range w.Days {
			if strings.EqualFold(day[:3], currentDay[:3]) {
				dayMatch = true
				break
			}
		}

		if dayMatch {
			start, _ := time.Parse("15:04", w.StartTime)
			startTime := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), 0, 0, loc)

			if t.Before(startTime) {
				return startTime
			}
		}

		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
	}

	return t
}
