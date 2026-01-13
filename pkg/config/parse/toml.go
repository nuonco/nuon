package parse

import (
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
)

// FileProcessor is a function to process config files before they're marshalled into a config struct and synced to the api.
type FileProcessor func(string, map[string]any) map[string]any

func parseTomlFile(rw io.ReadCloser, name string, out any, processor FileProcessor) error {

	tomlDec := toml.NewDecoder(rw)

	obj := make(map[string]interface{})
	err := tomlDec.Decode(&obj)
	if err != nil {
		return ParseErr{
			Filename:    name,
			Description: "unable to parse configuration file",
		}
	}

	// Skip files that are effectively empty (e.g., only comments)
	if len(obj) == 0 {
		return nil
	}

	obj = processor(name, obj)

	// go from map[string]interface{} => config.AppConfig
	mapDecCfg := config.DecoderConfig()
	mapDecCfg.Result = out
	mapDec, err := mapstructure.NewDecoder(mapDecCfg)
	if err != nil {
		return err
	}

	err = mapDec.Decode(obj)
	if err != nil {
		return ParseErr{
			Filename:    name,
			Description: "error decoding config",
			Err:         err,
		}
	}

	return nil
}
