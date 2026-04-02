package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReleaseWindowTestSuite struct {
	suite.Suite
}

func TestReleaseWindowSuite(t *testing.T) {
	suite.Run(t, new(ReleaseWindowTestSuite))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_MatchingDayAndTime() {
	w := &ReleaseWindow{
		Days:      []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Wednesday 2026-03-25 at 12:00 UTC
	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_WrongDay() {
	w := &ReleaseWindow{
		Days:      []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Saturday 2026-03-28 at 12:00 UTC
	t := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_BeforeStartTime() {
	w := &ReleaseWindow{
		Days:      []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Wednesday at 08:59 UTC
	t := time.Date(2026, 3, 25, 8, 59, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_AfterEndTime() {
	w := &ReleaseWindow{
		Days:      []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Wednesday at 17:01 UTC
	t := time.Date(2026, 3, 25, 17, 1, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_ExactlyAtStartTime() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_ExactlyAtEndTime() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// End time is exclusive
	t := time.Date(2026, 3, 25, 17, 0, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_InvalidTimezone_FallsBackToUTC() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "Invalid/Timezone",
	}

	// Wednesday at 12:00 UTC - should work because fallback is UTC
	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_InvalidStartTimeFormat() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "invalid",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_InvalidEndTimeFormat() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "invalid",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_WithTimezoneConversion() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "America/New_York",
	}

	// 14:00 UTC = 10:00 ET (within window)
	t := time.Date(2026, 3, 25, 14, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))

	// 13:00 UTC = 09:00 ET (exactly at start, should be open)
	t = time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))

	// 12:00 UTC = 08:00 ET (before window)
	t = time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.False(s.T(), w.IsOpen(t))
}

func (s *ReleaseWindowTestSuite) TestIsOpen_FullDayAbbreviations() {
	w := &ReleaseWindow{
		Days:      []string{"Wednesday"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.True(s.T(), w.IsOpen(t))
}

// NextOpenTime tests

func (s *ReleaseWindowTestSuite) TestNextOpenTime_CurrentlyOpen_ReturnsCurrent() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	assert.Equal(s.T(), t, w.NextOpenTime(t))
}

func (s *ReleaseWindowTestSuite) TestNextOpenTime_TodayBeforeStart() {
	w := &ReleaseWindow{
		Days:      []string{"Wed"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 7, 0, 0, 0, time.UTC)
	expected := time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC)
	assert.Equal(s.T(), expected, w.NextOpenTime(t))
}

func (s *ReleaseWindowTestSuite) TestNextOpenTime_TodayAfterEnd_AdvancesToNextDay() {
	w := &ReleaseWindow{
		Days:      []string{"Wed", "Thu"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Wednesday after end
	t := time.Date(2026, 3, 25, 18, 0, 0, 0, time.UTC)
	expected := time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	assert.Equal(s.T(), expected, w.NextOpenTime(t))
}

func (s *ReleaseWindowTestSuite) TestNextOpenTime_SkipsNonMatchingDays() {
	w := &ReleaseWindow{
		Days:      []string{"Mon"},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	// Wednesday after end - next Monday is 2026-03-30
	t := time.Date(2026, 3, 25, 18, 0, 0, 0, time.UTC)
	expected := time.Date(2026, 3, 30, 9, 0, 0, 0, time.UTC)
	assert.Equal(s.T(), expected, w.NextOpenTime(t))
}

func (s *ReleaseWindowTestSuite) TestNextOpenTime_EmptyDays_ReturnsFallback() {
	w := &ReleaseWindow{
		Days:      []string{},
		StartTime: "09:00",
		EndTime:   "17:00",
		Timezone:  "UTC",
	}

	t := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	// Should not hang; returns after 8 iterations
	result := w.NextOpenTime(t)
	assert.True(s.T(), result.After(t))
}
