package generics

import (
	"database/sql"
	"encoding/json"
	"strings"
)

type NullString struct {
	sql.NullString
}

func (s NullString) UnmarshalJSON(data []byte) error {
	s.String = strings.Trim(string(data), `"`)
	s.Valid = true
	return nil
}

func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		s.Valid = true
		s.String = ""
	}

	return json.Marshal(s.String)
}

func NewNullString(val string) NullString {
	return NullString{sql.NullString{
		String: val,
		Valid:  true,
	}}
}
