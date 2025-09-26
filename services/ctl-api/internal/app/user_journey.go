package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type UserJourneyStep struct {
	Name     string `json:"name" gorm:"column:name"`
	Title    string `json:"title" gorm:"column:title"`
	Complete bool   `json:"complete" gorm:"column:complete;default:false"`

	// Top-level completion tracking fields
	CompletedAt      *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	CompletionMethod string     `json:"completion_method,omitempty" gorm:"column:completion_method"`
	CompletionSource string     `json:"completion_source,omitempty" gorm:"column:completion_source"`

	// Flexible metadata for business data
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"column:metadata;type:jsonb"`
}

type UserJourney struct {
	Name  string            `json:"name" gorm:"column:name"`
	Title string            `json:"title" gorm:"column:title"`
	Steps []UserJourneyStep `json:"steps" gorm:"column:steps;type:jsonb"`
}

// UserJourneys represents a slice of UserJourney that can be stored in JSONB
type UserJourneys []UserJourney

// Scan implements the database/sql.Scanner interface.
func (uj *UserJourneys) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		*uj = nil
		return nil
	case []byte:
		// First, try to unmarshal as an array (expected format)
		if err := json.Unmarshal(v, uj); err != nil {
			// If that fails, try to unmarshal as a single object and wrap it in a slice
			var single UserJourney
			if singleErr := json.Unmarshal(v, &single); singleErr != nil {
				// If both fail, return the original array error with more context
				return errors.Wrapf(err, "unable to scan user journeys as array, single object also failed: %v", singleErr)
			}
			// Successfully unmarshaled as single object, wrap in slice
			*uj = UserJourneys{single}
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (uj UserJourneys) Value() (driver.Value, error) {
	if uj == nil {
		return nil, nil
	}
	return json.Marshal(uj)
}

func (UserJourneys) GormDataType() string {
	return "jsonb"
}
