package converters

import (
	"time"

	"google.golang.org/genproto/googleapis/type/datetime"
)

func TimeToDatetime(ts time.Time) *datetime.DateTime {
	ts = ts.In(time.UTC)
	return &datetime.DateTime{
		Year:       int32(ts.Year()),
		Month:      int32(ts.Month()),
		Day:        int32(ts.Day()),
		Hours:      int32(ts.Hour()),
		Minutes:    int32(ts.Minute()),
		Seconds:    int32(ts.Second()),
		Nanos:      int32(ts.Nanosecond()),
		TimeOffset: &datetime.DateTime_UtcOffset{},
	}
}
