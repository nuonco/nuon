package generics

import (
	"database/sql"
	"encoding/json"
	"time"
)

type NullTime struct {
	sql.NullTime
}

func (dt NullTime) Empty() bool {
	return dt.Time.IsZero()
}

func (dt NullTime) ValueTime() time.Time {
	return dt.Time
}

func (dt NullTime) ValueOrDefault(def time.Time) time.Time {
	if dt.Empty() {
		return def
	}

	return dt.ValueTime()
}

func (dt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		dt.Valid = false
		dt.Time = time.Time{}
		return nil
	}

	parsedTime, err := time.Parse(`"2006-01-02T15:04:05Z07:00"`, string(data))
	if err != nil {
		return err
	}
	dt.Time = parsedTime
	dt.Valid = true
	return nil
}

func (dt NullTime) MarshalJSON() ([]byte, error) {
	if !dt.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(dt.Time.Format(`2006-01-02T15:04:05Z07:00`))
}

func NewNullTime(val time.Time) NullTime {
	return NullTime{sql.NullTime{
		Time:  val,
		Valid: !val.IsZero(),
	}}
}
