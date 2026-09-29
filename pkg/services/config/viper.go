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
	ConfigTagName  = "config"
	TagValueSquash = ",squash"

	loaderDefaultSize = 10
)

var (
	mapStringStringType = reflect.TypeOf(map[string]string{})
	unmarshalerType     = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

type loader struct {
	defaults        map[string]interface{}
	file            string
	additionalPaths []string
	config          *config
	envVarPrefix    string
}

type config struct {
	*viper.Viper
	secure map[string]struct{}
}

func NewLoader(paths ...string) Loader {
	loader := new(loader)
	loader.defaults = make(map[string]interface{}, loaderDefaultSize)
	loader.additionalPaths = paths
	return loader
}

func NewFileLoader(file string) Loader {
	loader := new(loader)
	loader.defaults = make(map[string]interface{}, loaderDefaultSize)
	loader.file = file
	return loader
}

func (l *loader) RegisterDefault(key string, value interface{}) {
	l.defaults[key] = value
}

func (l *loader) RegisterDefaults(kvs map[string]interface{}) {
	for key, value := range kvs {
		l.RegisterDefault(key, value)
	}
}

func (l *loader) SetEnvPrefix(prefix string) {
	l.envVarPrefix = prefix
}

//nolint:unparam // NOTE(jdt): this is inherited.
func (l *loader) newConfig(flags *pflag.FlagSet) (*config, error) {
	c := new(config)
	c.Viper = viper.New()
	c.secure = make(map[string]struct{}, loaderDefaultSize)

	c.SetEnvPrefix(l.envVarPrefix)
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.AutomaticEnv()

	_, file := path.Split(os.Args[0])

	c.SetConfigName("config")
	c.AddConfigPath(fmt.Sprintf("/etc/%s/", file))
	c.AddConfigPath(fmt.Sprintf("$HOME/.%s", file))
	c.AddConfigPath(".")
	c.SetConfigFile(l.file)
	for _, additionalPath := range l.additionalPaths {
		c.AddConfigPath(additionalPath)
	}

	for k, v := range l.defaults {
		c.SetDefault(k, v)
	}

	if flags != nil {
		_ = c.BindPFlags(flags)
	}

	_ = c.ReadInConfig()
	return c, nil
}

func (l *loader) Load(flags *pflag.FlagSet) (Config, error) {
	config, err := l.newConfig(flags)
	if err != nil {
		return nil, err
	}
	l.config = config
	return config, nil
}

func (l *loader) LoadInto(flags *pflag.FlagSet, to interface{}) error {
	config, err := l.Load(flags)
	if err != nil {
		return err
	}

	return config.Bind(to)
}

func (l *loader) LoadedConfig() Config {
	return l.config
}

func (c *config) Bind(to interface{}) error {
	err := c.bindEnvVars(reflect.TypeOf(to), "")
	if err != nil {
		return err
	}

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

func (c *config) unmarshal(to reflect.Type, str string) interface{} {
	obj := reflect.New(to).Interface()
	if unmarshaler, ok := obj.(Unmarshaler); ok {
		unmarshaler.UnmarshalConfig(str)
	}
	return obj
}

func (c *config) bindEnvVars(to reflect.Type, prefix string) error {
	if to == nil {
		return nil
	}

	if to.Kind() == reflect.Ptr {
		to = to.Elem()
	}

	for i := 0; i < to.NumField(); i++ {
		field := to.Field(i)
		if field.PkgPath != "" {
			continue
		}

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

		var err error
		if field.Type.Kind() == reflect.Struct {
			err = c.bindEnvVars(field.Type, tag)
		} else if field.Type.Kind() == reflect.Ptr {
			err = c.bindEnvVars(field.Type.Elem(), tag)
		} else if !strings.HasSuffix(tag, ".") {
			err = c.BindEnv(tag)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *config) GetStringSlice(key string) []string {
	result := make([]string, 0)
	for _, entry := range c.Viper.GetStringSlice(key) {
		csvReader := csv.NewReader(strings.NewReader(entry))
		split, err := csvReader.Read()
		if err != nil {
			continue
		}
		for _, part := range split {
			if part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}

func (c *config) GetStringMapString(key string) map[string]string {
	result := c.Viper.GetStringMapString(key)
	if c.IsSet(key) && len(result) == 0 {
		return c.parseMap(c.GetString(key))
	}
	return result
}

func (c *config) WriteTo(w io.Writer) (int64, error) {
	settings := c.AllSettings()
	c.maskSecure(settings, "")

	b, err := json.Marshal(settings)
	if err != nil {
		return -1, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

func (c *config) SetSecure(key string) {
	c.secure[key] = struct{}{}
}

func (c *config) Secure(key string) bool {
	_, ok := c.secure[key]
	return ok
}

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
