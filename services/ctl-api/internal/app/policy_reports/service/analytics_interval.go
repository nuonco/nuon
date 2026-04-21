package service

import (
	"fmt"
	"time"
)

type timeInterval struct {
	Label  string
	chExpr string // ClickHouse expression template with %s placeholder for column name
}

func (t timeInterval) BucketExpr(column string) string {
	return fmt.Sprintf(t.chExpr, column)
}

func intervalForRange(start, end time.Time) timeInterval {
	d := end.Sub(start)
	switch {
	case d <= 24*time.Hour:
		return timeInterval{"hour", "toStartOfHour(%s)"}
	case d <= 7*24*time.Hour:
		return timeInterval{"6h", "toStartOfInterval(%s, INTERVAL 6 HOUR)"}
	case d <= 30*24*time.Hour:
		return timeInterval{"day", "toStartOfDay(%s)"}
	case d <= 90*24*time.Hour:
		return timeInterval{"week", "toStartOfWeek(%s)"}
	default:
		return timeInterval{"month", "toStartOfMonth(%s)"}
	}
}
