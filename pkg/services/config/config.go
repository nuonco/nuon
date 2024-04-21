package config

import (
	"io"
	"time"

	"github.com/spf13/pflag"
)

// Config provides access to application configuration information
type Config interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool
	Set(key string, value interface{})
	Bind(to interface{}) error
	Debug()
	SetSecure(key string)
	WriteTo(w io.Writer) (int64, error)
}

// Loader loads configurations from multiple sources
type Loader interface {
	RegisterDefault(key string, value interface{})
	RegisterDefaults(kvs map[string]interface{})
	SetEnvPrefix(string)
	Load(flags *pflag.FlagSet) (Config, error)
	LoadInto(flags *pflag.FlagSet, to interface{}) error
	LoadedConfig() Config
}

// Unmarshaler describes a custom Unmarshal function that can be used to bind
// config values
type Unmarshaler interface {
	UnmarshalConfig(value string)
}

// loader is a global Loader instance
var stdLoader = NewLoader()

// RegisterDefault registers a default value for a configuration key
func RegisterDefault(key string, value interface{}) {
	stdLoader.RegisterDefault(key, value)
}

// RegisterDefaults register the set of default key/value pairs
func RegisterDefaults(kvs map[string]interface{}) {
	stdLoader.RegisterDefaults(kvs)
}

// SetEnvPrefix sets the prefix to env vars that are searched
func SetEnvPrefix(prefix string) {
	stdLoader.SetEnvPrefix(prefix)
}

// Load load the configuration using the standard Loader
func Load(flags *pflag.FlagSet) (Config, error) {
	return stdLoader.Load(flags)
}

// LoadInto load the configuration into the supplied struct using the standard
// Loader
func LoadInto(flags *pflag.FlagSet, to interface{}) error {
	return stdLoader.LoadInto(flags, to)
}
