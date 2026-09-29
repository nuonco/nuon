package parse

import (
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
)

type FileProcessor func(string, map[string]any) map[string]any

func parseTomlFile(rw io.ReadCloser, name string, out any, processor FileProcessor, rootDir string) error {

	tomlDec := toml.NewDecoder(rw)

	obj := make(map[string]interface{})
	err := tomlDec.Decode(&obj)
	if err != nil {
		return ParseErr{
			Filename:    name,
			Description: "unable to parse configuration file",
		}
	}

	if len(obj) == 0 {
		return nil
	}

	obj = processor(name, obj)

	mapDecCfg := config.DecoderConfig(config.WithRootDir(rootDir))
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
