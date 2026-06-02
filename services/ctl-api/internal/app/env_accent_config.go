package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// EnvAccentColor is the constrained set of accent themes installs can be
// painted with in the dashboard. Mapped 1:1 onto the existing TTheme tokens
// on the frontend, so we never introduce new color tokens.
type EnvAccentColor string

const (
	EnvAccentColorError   EnvAccentColor = "error"   // red — production
	EnvAccentColorWarn    EnvAccentColor = "warn"    // amber — staging
	EnvAccentColorSuccess EnvAccentColor = "success" // green — dev
	EnvAccentColorInfo    EnvAccentColor = "info"    // blue — qa
	EnvAccentColorBrand   EnvAccentColor = "brand"   // violet — misc
	EnvAccentColorNeutral EnvAccentColor = "neutral" // grey — default
)

// ValidEnvAccentColors returns the allowlist of accent colors. Validation
// against this set happens at the API boundary so the column never holds
// an arbitrary string.
func ValidEnvAccentColors() []EnvAccentColor {
	return []EnvAccentColor{
		EnvAccentColorError,
		EnvAccentColorWarn,
		EnvAccentColorSuccess,
		EnvAccentColorInfo,
		EnvAccentColorBrand,
		EnvAccentColorNeutral,
	}
}

// EnvAccentConfig is the per-org mapping that drives install environment
// indicators in the dashboard. Operators set one label key (default `env`)
// and map known values to colors. Installs whose labels match render with
// the matching accent everywhere (status bar chip, page stripe, table row).
//
// Stored as JSONB on Org.env_accent_config.
type EnvAccentConfig struct {
	LabelKey string                    `json:"label_key"`
	Values   map[string]EnvAccentColor `json:"values"`
}

// DefaultEnvAccentConfig is the out-of-the-box mapping used for new orgs
// and as a safe fallback if the column is null.
func DefaultEnvAccentConfig() EnvAccentConfig {
	return EnvAccentConfig{
		LabelKey: "env",
		Values: map[string]EnvAccentColor{
			"production":  EnvAccentColorError,
			"prod":        EnvAccentColorError,
			"staging":     EnvAccentColorWarn,
			"stage":       EnvAccentColorWarn,
			"qa":          EnvAccentColorInfo,
			"dev":         EnvAccentColorSuccess,
			"development": EnvAccentColorSuccess,
		},
	}
}

// Validate enforces label_key non-empty (when any values are set) and a
// closed enum on the color side. Returns nil for a fully empty config.
func (c EnvAccentConfig) Validate() error {
	if len(c.Values) == 0 {
		return nil
	}
	if c.LabelKey == "" {
		return errors.New("label_key is required when values are set")
	}
	allowed := map[EnvAccentColor]struct{}{}
	for _, color := range ValidEnvAccentColors() {
		allowed[color] = struct{}{}
	}
	for value, color := range c.Values {
		if value == "" {
			return errors.New("value entries must have a non-empty label value")
		}
		if _, ok := allowed[color]; !ok {
			return fmt.Errorf("invalid accent color %q for value %q", color, value)
		}
	}
	return nil
}

// Scan implements sql.Scanner for JSONB deserialization.
func (c *EnvAccentConfig) Scan(value interface{}) error {
	if value == nil {
		*c = EnvAccentConfig{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for EnvAccentConfig")
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for JSONB serialization.
func (c EnvAccentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// GormDataType returns the GORM data type for this field.
func (EnvAccentConfig) GormDataType() string {
	return "jsonb"
}
