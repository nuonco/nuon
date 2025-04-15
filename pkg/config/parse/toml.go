package parse

import (
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"

	"github.com/powertoolsdev/mono/pkg/config"
)

func parseTomlFile(rw io.ReadCloser, name string, out any) error {
	tomlDec := toml.NewDecoder(rw)
	tomlDec.SetTagName("mapstructure")

	obj := make(map[string]interface{})
	err := tomlDec.Decode(&obj)
	if err != nil {
		return ParseErr{
			Description: "unable to parse configuration file",
		}
	}

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
