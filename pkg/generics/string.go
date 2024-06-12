package generics

import (
	"database/sql"
	"encoding/json"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func FirstNonEmptyString(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}

	return ""
}

func DisplayName(val string) string {
	str := strings.ReplaceAll(string(val), "_", " ")

	caser := cases.Title(language.English)
	str = caser.String(str)
	return str
}

type NullString struct {
	sql.NullString
}

func (s *NullString) UnmarshalJSON(data []byte) error {
	s.String = strings.Trim(string(data), `"`)
	s.Valid = true
	return nil
}

func (s *NullString) MarshalJSON() ([]byte, error) {
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
