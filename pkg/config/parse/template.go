package parse

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

// NOTE(jm): we have to be careful about just using template/text here, because it will mark all of our user config
// values as empty or errors.
//
// Thus, we have a very simple substition in our config file, that just reads the top level keys and uses
// bytes.Replace
const (
	tmplSubKey string = "{{.%s}}"
)

type BaseConfig struct {
	ConfigVars map[string]interface{} `toml:"config_vars"`
}

func Template(byts []byte) ([]byte, error) {
	var cfg BaseConfig
	if err := toml.Unmarshal(byts, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse toml min config: %w", err)
	}

	for key, val := range cfg.ConfigVars {
		replaceKey := fmt.Sprintf(tmplSubKey, key)
		replaceVal := fmt.Sprintf("%s", val)
		byts = bytes.ReplaceAll(byts, []byte(replaceKey), []byte(replaceVal))
	}

	return byts, nil
}
