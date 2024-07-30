package config

import (
	"bytes"

	"github.com/pelletier/go-toml"
)

func (a *AppConfig) ToTOML() ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetTagName("mapstructure")

	err := enc.Encode(a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
