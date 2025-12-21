package queue

import (
	"strings"
	"time"
)

type ReleaseWindow struct {
	// Days of the week (Mon, Tue, Wed, Thu, Fri, Sat, Sun)
	Days []string
	// StartTime in HH:MM format
	StartTime string
	// EndTime in HH:MM format
	EndTime string
	// Timezone (e.g. "America/New_York")
	Timezone string
}

func (w *ReleaseWindow) IsOpen(t time.Time) bool {
	loc, err := time.LoadLocation(w.Timezone)
	if err != nil {
		// default to UTC if timezone is invalid
		loc = time.UTC
	}
	t = t.In(loc)

	// check day
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

	// parse times
	start, err := time.Parse("15:04", w.StartTime)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04", w.EndTime)
	if err != nil {
		return false
	}

	// create time objects for today with start/end times
	startTime := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), 0, 0, loc)
	endTime := time.Date(t.Year(), t.Month(), t.Day(), end.Hour(), end.Minute(), 0, 0, loc)

	return (t.Equal(startTime) || t.After(startTime)) && t.Before(endTime)
}

// NextOpenTime returns the next time the window opens.
// If the window is currently open, it returns the current time.
func (w *ReleaseWindow) NextOpenTime(t time.Time) time.Time {
	if w.IsOpen(t) {
		return t
	}

	loc, err := time.LoadLocation(w.Timezone)
	if err != nil {
		loc = time.UTC
	}
	t = t.In(loc)

	// prevent infinite loop in case of bad config
	for i := 0; i < 8; i++ {
		// check if today is a valid day
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

			// if we are before the start time today, then this is the next open time
			if t.Before(startTime) {
				return startTime
			}
		}

		// advance to next day
		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
	}

	return t
}
