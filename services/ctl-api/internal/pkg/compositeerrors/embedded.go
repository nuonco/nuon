package compositeerrors

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// CompositeErrorData is the JSONB GORM column attached to owner rows. It
// captures a typed CompositeError's payload along with its headline message
// and structured sections, all frozen at write time.
//
// To use, add this to a GORM model struct:
//
//	CompositeError *compositeerrors.CompositeErrorData `json:"composite_error,omitempty" gorm:"type:jsonb"`
type CompositeErrorData struct {
	Type     Type            `json:"type"`
	Severity Severity        `json:"severity"`
	Message  string          `json:"message"`
	Sections []Section       `json:"sections,omitempty"`
	Data     json.RawMessage `json:"data"`
}

// New constructs a CompositeErrorData from a typed CompositeError. The
// implementation's data, headline message, and sections are captured at this
// point, all are frozen on the resulting record.
func New(e CompositeError) *CompositeErrorData {
	data, _ := json.Marshal(e)
	return &CompositeErrorData{
		Type:     e.Type(),
		Severity: e.Severity(),
		Message:  e.Error(),
		Sections: e.Sections(),
		Data:     data,
	}
}

// Scan implements database/sql.Scanner.
func (c *CompositeErrorData) Scan(value any) error {
	if value == nil {
		*c = CompositeErrorData{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("compositeerrors: cannot scan type %T", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*c = CompositeErrorData{}
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer.
func (c *CompositeErrorData) Value() (driver.Value, error) {
	if c == nil || c.Type == "" {
		return nil, nil
	}
	return json.Marshal(c)
}

// GormDataType tells GORM to use a jsonb column.
func (CompositeErrorData) GormDataType() string { return "jsonb" }
