package registry

import (
	"fmt"

	"github.com/distribution/distribution/v3/configuration"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
)

func (r *Registry) getConfig(port int) *configuration.Configuration {
	cfg := &configuration.Configuration{
		Storage: make(map[string]configuration.Parameters),
	}

	cfg.Log.Level = "info"
	cfg.HTTP.Addr = fmt.Sprintf(":%d", port)
	cfg.HTTP.Host = "localhost"

	cfg.Storage["filesystem"] = configuration.Parameters{
		"rootdirectory": r.cfg.RegistryDir,
	}

	return cfg
}
