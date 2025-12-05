package parse

import (
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"

	"github.com/powertoolsdev/mono/pkg/config"
)

// FileProcessor is a function to process config files before they're marshalled into a config struct and synced to the api.
type FileProcessor func(string, map[string]any) map[string]any

func parseTomlFile(rw io.ReadCloser, name string, out any, processor FileProcessor) error {

	tomlDec := toml.NewDecoder(rw)
	tomlDec.SetTagName("mapstructure")

	obj := make(map[string]interface{})
	err := tomlDec.Decode(&obj)
	if err != nil {
		return ParseErr{
			Description: "unable to parse configuration file",
		}
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
			Description: "error decoding config",
			Err:         err,
		}
	}

	return nil
}
