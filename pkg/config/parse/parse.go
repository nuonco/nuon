package parse

import (
	"bytes"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
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

	// go from toml -> map[string]interface{}
	buf := bytes.NewReader(byts)
	tomlDec := toml.NewDecoder(buf)
	tomlDec.SetTagName("mapstructure")

	obj := make(map[string]interface{})
	err = tomlDec.Decode(&obj)
	if err != nil {
		return nil, ParseErr{
			Description: "unable to parse configuration file",
		}
	}

	// go from map[string]interface{} => config.AppConfig
	var cfg config.AppConfig
	mapDecCfg := config.DecoderConfig()
	mapDecCfg.Result = &cfg
	mapDec, err := mapstructure.NewDecoder(mapDecCfg)
	if err != nil {
		return nil, err
	}

	err = mapDec.Decode(obj)
	if err != nil {
		return nil, ParseErr{
			Description: "error decoding config",
			Err:         err,
		}
	}

	err = cfg.Parse()
	if err != nil {
		return nil, ParseErr{
			Description: "error parsing config",
			Err:         err,
		}
	}

	return &cfg, nil
}
