package config

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// ConfigTagName is the name of the field tag to use for config values
	ConfigTagName = "config"
	// TagValueSquash is the tag value used when squashing embedded structs
	TagValueSquash = ",squash"

	loaderDefaultSize = 10
)

var (
	// mapStringStringType is the type associated with a map[string]string
	mapStringStringType = reflect.TypeOf(map[string]string{})
	// unmarshalerType is the type associated with a custom Unmarshaler
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

// loader implements the loader interface using Viper to manage the underlying
// configuration loading
type loader struct {
	defaults        map[string]interface{}
	file            string
	additionalPaths []string
	config          *config
	envVarPrefix    string
}

// config implments the Config interface backed by Viper
type config struct {
	*viper.Viper
	secure map[string]struct{}
}

// NewLoader creates a new config Loader instance
func NewLoader(paths ...string) Loader {
	loader := new(loader)
	loader.defaults = make(map[string]interface{}, loaderDefaultSize)
	loader.additionalPaths = paths
	return loader
}

// NewFileLoader creates a new config Loader instance with the supplied config
// file
func NewFileLoader(file string) Loader {
	loader := new(loader)
	loader.defaults = make(map[string]interface{}, loaderDefaultSize)
	loader.file = file
	return loader
}

// RegisterDefault registers the supplied value as the default for the supplied
// key. When no config value is found for the given key in any of the sources,
// this value will be returned
func (l *loader) RegisterDefault(key string, value interface{}) {
	l.defaults[key] = value
}

// RegisterDefaults registers the supplied set of key / value pairs as defaults.
// When no config value is found for a given key in any of the sources, the
// corresponding value in the map will be returned
func (l *loader) RegisterDefaults(kvs map[string]interface{}) {
	for key, value := range kvs {
		l.RegisterDefault(key, value)
	}
}

// SetEnvPrefix sets the prefix that the underlying implementation uses when
// searching for env vars to bind to configuration.
func (l *loader) SetEnvPrefix(prefix string) {
	l.envVarPrefix = prefix
}

// newConfig will initalize and return a new Config object backed by Viper
//
//nolint:unparam // NOTE(jdt): this is inherited.
func (l *loader) newConfig(flags *pflag.FlagSet) (*config, error) {
	c := new(config)
	c.Viper = viper.New()
	c.secure = make(map[string]struct{}, loaderDefaultSize)

	// Set up ENV loading
	c.Viper.SetEnvPrefix(l.envVarPrefix)
	c.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.Viper.AutomaticEnv()

	// Extract the filename to use in building paths
	_, file := path.Split(os.Args[0])

	// Set up File loading
	c.Viper.SetConfigName("config")
	c.Viper.AddConfigPath(fmt.Sprintf("/etc/%s/", file))
	c.Viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", file))
	c.Viper.AddConfigPath(".")
	c.Viper.SetConfigFile(l.file)
	for _, additionalPath := range l.additionalPaths {
		c.Viper.AddConfigPath(additionalPath)
	}

	// Apply the defaults
	for k, v := range l.defaults {
		c.Viper.SetDefault(k, v)
	}

	// Bind the command line flags
	if flags != nil {
		// ignore invalid flags in case of error:
		_ = c.Viper.BindPFlags(flags)
	}

	// Read in the file config
	_ = c.Viper.ReadInConfig()
	return c, nil
}

// Load will load the configuration.  It will aggregate values found in the
// arguments, environment, and any configuration file in the specified search
// paths together applying the following precedence:
//
// -flag
// -env
// -config
// -default
func (l *loader) Load(flags *pflag.FlagSet) (Config, error) {
	config, err := l.newConfig(flags)
	if err != nil {
		return nil, err
	}
	l.config = config
	return config, nil
}

// LoadInfo will load the specified configuration via the Load function and
// Bind it into the specified struct.
func (l *loader) LoadInto(flags *pflag.FlagSet, to interface{}) error {
	// Load the Config
	config, err := l.Load(flags)
	if err != nil {
		return err
	}

	// Bind it to the supplied struct
	return config.Bind(to)
}

// LoadedConfig returns the loaded config
func (l *loader) LoadedConfig() Config {
	return l.config
}

// Bind will Bind the values in the configuration to the corresponding struct
// fields. Fields can be tagged with `config` to specify what configuration
// value should be used
func (c *config) Bind(to interface{}) error {
	// Bind the possible ENV vars
	err := c.bindEnvVars(reflect.TypeOf(to), "")
	if err != nil {
		return err
	}

	// Unmarshal the config into the supplied struct, using `config` as the tag
	// name. By default, Viper does not support binding of maps from ENV vars,
	// so add a custom hook to handle map encoded strings
	return c.Unmarshal(&to, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = ConfigTagName
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			func(from reflect.Type, to reflect.Type, value interface{}) (interface{}, error) {
				if to == mapStringStringType {
					if str, ok := value.(string); ok {
						return c.parseMap(str), nil
					}
				} else if reflect.PtrTo(to).Implements(unmarshalerType) {
					if str, ok := value.(string); ok {
						return c.unmarshal(to, str), nil
					}
				}
				return value, nil
			},
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	})
}

// parseMap parses the map out of the string. The expected format is:
//
// key:value,key:value
func (c *config) parseMap(str string) map[string]string {
	pairs := strings.Split(str, ",")
	value := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		indx := strings.Index(pair, ":")
		if indx > -1 {
			value[strings.ToLower(pair[0:indx])] = pair[indx+1:]
		} else {
			value[strings.ToLower(pair)] = ""
		}
	}
	return value
}

// unmarshal calls the custom Unmarshal function on the to type to unmarshal the
// specified string
func (c *config) unmarshal(to reflect.Type, str string) interface{} {
	obj := reflect.New(to).Interface()
	if unmarshaler, ok := obj.(Unmarshaler); ok {
		unmarshaler.UnmarshalConfig(str)
	}
	return obj
}

// bindEnvVars will introspect the supplied type and bind the equivalent ENV
// vars so they are available when binding the struct
func (c *config) bindEnvVars(to reflect.Type, prefix string) error {
	// If we were passed nil there is nothing to do
	if to == nil {
		return nil
	}

	// TypeOf returns the reflection Type that represents the dynamic type of
	// variable
	if to.Kind() == reflect.Ptr {
		to = to.Elem()
	}

	// Iterate over all available fields and read the tag value
	for i := 0; i < to.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := to.Field(i)
		if field.PkgPath != "" {
			continue
		}

		// Get the field tag value, appending the prefix if one was provided
		tag, options := parseTag(field.Tag.Get(ConfigTagName))
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		if options.Contains("squash") {
			tag = prefix
		} else if prefix != "" {
			tag = prefix + "." + tag
		}
		if options.Contains("secure") {
			c.SetSecure(tag)
		}

		// If its a nested Struct, recursively bind
		var err error
		if field.Type.Kind() == reflect.Struct {
			err = c.bindEnvVars(field.Type, tag)
		} else if field.Type.Kind() == reflect.Ptr {
			err = c.bindEnvVars(field.Type.Elem(), tag)
		} else if !strings.HasSuffix(tag, ".") {
			err = c.Viper.BindEnv(tag)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// GetStringSlice works around an issue with parsing string slices
// https://github.com/spf13/viper/issues/380
func (c *config) GetStringSlice(key string) []string {
	result := make([]string, 0)
	for _, entry := range c.Viper.GetStringSlice(key) {
		csvReader := csv.NewReader(strings.NewReader(entry))
		split, err := csvReader.Read()
		if err != nil {
			continue
		}
		for _, part := range split {
			if part != "" { // Don't add empty values
				result = append(result, part)
			}
		}
	}
	return result
}

// GetStringMapString works around an issue with parsing string maps
func (c *config) GetStringMapString(key string) map[string]string {
	result := c.Viper.GetStringMapString(key)
	if c.Viper.IsSet(key) && len(result) == 0 {
		return c.parseMap(c.Viper.GetString(key))
	}
	return result
}

// WriteTo dumps the config to the supplied writer
func (c *config) WriteTo(w io.Writer) (int64, error) {
	// Get all the settings, masking those that are secure
	settings := c.Viper.AllSettings()
	c.maskSecure(settings, "")

	// Marshall the settings to JSON
	b, err := json.Marshal(settings)
	if err != nil {
		return -1, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

// SetSecure masks a field when writing
func (c *config) SetSecure(key string) {
	c.secure[key] = struct{}{}
}

// Secure returns if the supplied key is secure
func (c *config) Secure(key string) bool {
	_, ok := c.secure[key]
	return ok
}

// maskSecure masks the fields that have been marked as secure either via the
// config tag or an explicit call to SetSecure
func (c *config) maskSecure(settings map[string]interface{}, prefix string) {
	for k, v := range settings {
		var path string
		if len(prefix) > 0 {
			path = prefix + "." + k
		} else {
			path = k
		}
		settings[k] = c.mask(v, path)
	}
}

// mask masks the supplied value at the given path
func (c *config) mask(v interface{}, path string) interface{} {
	l := 10
	if c.Secure(path) {
		return strings.Repeat("*", l)
	}
	switch obj := v.(type) {
	case map[string]interface{}:
		c.maskSecure(obj, path)
	case []interface{}:
		for i, item := range obj {
			obj[i] = c.mask(item, path)
		}
	}
	return v
}
