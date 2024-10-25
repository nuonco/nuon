package registry

import (
	"github.com/distribution/distribution/v3/configuration"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
)

// NOTE(jm): this represents a minimal config to run the registry in process, as essentially a local cache with no
// characteristics
func (r *Registry) getConfig() *configuration.Configuration {
	cfg := &configuration.Configuration{
		Storage: make(map[string]configuration.Parameters),
	}

	// basic parameters for listening/logging
	cfg.Log.Level = "info"
	cfg.HTTP.Addr = ":5000"
	cfg.HTTP.Host = "localhost"

	// an (albeit partially outdated) configuration exists here -
	// https://github.com/GerritForge/docker-registry/blob/master/docs/configuration.md
	//
	// NOTE(jm): eventually, we may consider using S3 as a registry backend, or some other type of non-ephemeral
	// backend.
	cfg.Storage["filesystem"] = configuration.Parameters{
		"rootdirectory": r.cfg.RegistryDir,
	}

	return cfg
}
