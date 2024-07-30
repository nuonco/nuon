package parse

import (
	"bytes"

	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml"

	"github.com/powertoolsdev/mono/pkg/config"
)

type ParseConfig struct {
	Filename string
	Bytes    []byte

	BackendType config.BackendType
	V           *validator.Validate
	Template    bool
	Context     config.ConfigContext
}

func Parse(parseCfg ParseConfig) (*config.AppConfig, error) {
	var cfg config.AppConfig

	if parseCfg.Filename != "" {
		byts, err := ReadFile(parseCfg.Filename)
		if err != nil {
			return nil, err
		}

		if len(string(byts)) == 0 {
			return nil, ParseErr{
				Description: "config file is empty",
			}
		}

		parseCfg.Bytes = byts
	}

	byts, err := Template(parseCfg.Bytes)
	if err != nil {
		return nil, ParseErr{
			Description: "unable to template values in config file",
			Err:         err,
		}
	}

	buf := bytes.NewReader(byts)
	dec := toml.NewDecoder(buf)
	dec.SetTagName("mapstructure")

	err = dec.Decode(&cfg)
	if err != nil {
		return nil, ParseErr{
			Description: "unable to parse configuration file",
		}
	}

	if err := cfg.Parse(parseCfg.Context); err != nil {
		return nil, ParseErr{
			Description: "unable to parse config",
			Err:         err,
		}
	}

	return &cfg, nil
}
