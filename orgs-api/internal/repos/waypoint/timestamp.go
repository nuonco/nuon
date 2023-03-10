package waypoint

import (
	"google.golang.org/genproto/googleapis/type/datetime"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TimestampToDatetime(pbTS *timestamppb.Timestamp) *datetime.DateTime {
	ts := pbTS.AsTime()
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
