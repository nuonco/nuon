package parse

import (
	"bytes"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
)

type ParseConfig struct {
	Filename string
	Dirname  string

	Bytes []byte

	BackendType   config.BackendType
	V             *validator.Validate
	Template      bool
	FileProcessor func(name string, obj map[string]any) map[string]any
}

func Parse(parseCfg ParseConfig) (*config.AppConfig, error) {
	if parseCfg.Filename != "" {
		byts, err := ReadFile(parseCfg.Filename)
		if err != nil {
			return nil, err
		}

		if len(string(byts)) == 0 {
			return nil, ParseErr{
				Filename:    parseCfg.Filename,
				Description: "config file is empty",
			}
		}

		parseCfg.Bytes = byts
	}

	byts, err := Template(parseCfg.Bytes)
	if err != nil {
		return nil, ParseErr{
			Filename:    parseCfg.Filename,
			Description: "unable to template values in config file",
			Err:         err,
		}
	}

	// go from toml -> map[string]interface{}
	buf := bytes.NewReader(byts)
	tomlDec := toml.NewDecoder(buf)

	obj := make(map[string]interface{})
	err = tomlDec.Decode(&obj)
	if err != nil {
		return nil, ParseErr{
			Filename:    parseCfg.Filename,
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
			Filename:    parseCfg.Filename,
			Description: "error decoding config",
			Err:         err,
		}
	}

	err = cfg.Parse()
	if err != nil {
		return nil, ParseErr{
			Filename:    parseCfg.Filename,
			Description: "error parsing config",
			Err:         err,
		}
	}

	return &cfg, nil
}
