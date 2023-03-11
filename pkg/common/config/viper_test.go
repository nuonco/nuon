package config

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CustomType is an enum determining the type of service to generate
type CustomType int8

// UnmarshalConfig unmarshals a config value string to the associated interface
// type
func (c *CustomType) UnmarshalConfig(value string) {
	switch value {
	case "one":
		*c = CustomTypeOne
	case "two":
		*c = CustomTypeTwo
	}
}

const (
	CustomTypeOne CustomType = iota
	CustomTypeTwo
)

type TestConfig struct {
	String     string            `config:"string"`
	Int        int               `config:"int"`
	Float      float64           `config:"float"`
	Map        map[string]string `config:"map"`
	Slice      []string          `config:"slice"`
	Duration   time.Duration     `config:"duration"`
	Custom     CustomType        `config:"custom"`
	Password   string            `config:"password"`
	MissingTag string
}

type TestSecureConfig struct {
	String     string            `config:"string"`
	Int        int               `config:"int"`
	Float      float64           `config:"float"`
	Map        map[string]string `config:"map,secure"`
	Slice      []string          `config:"slice,secure"`
	Duration   time.Duration     `config:"duration"`
	Custom     CustomType        `config:"custom"`
	Password   string            `config:"password,secure"`
	MissingTag string
	Nested     struct {
		Foo string `config:"foo,secure"`
	} `config:"nested"`
}

type TestNestedConfig struct {
	String    string      `config:"string"`
	NestedPtr *TestConfig `config:"nestedptr"`
	Nested    TestConfig  `config:"nested"`
}

func TestViper(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		t.Run("file", func(t *testing.T) {
			t.Run("json", func(t *testing.T) {
				loader := NewLoader("./testdata/json")

				cfg, err := loader.Load(nil)
				require.NoError(t, err)
				assertConfig(t, cfg, nil)
			})

			t.Run("yaml", func(t *testing.T) {
				loader := NewLoader("./testdata/yaml")

				cfg, err := loader.Load(nil)
				require.NoError(t, err)
				assertConfig(t, cfg, nil)
			})

			t.Run("explicit", func(t *testing.T) {
				loader := NewFileLoader("./testdata/yaml/config.yaml")

				config, err := loader.Load(nil)
				require.NoError(t, err)
				assertConfig(t, config, nil)
			})

			t.Run("override", func(t *testing.T) {
				t.Run("env", func(t *testing.T) {
					os.Setenv("STRING", "env_value")

					loader := NewLoader("./testdata/yaml")

					config, err := loader.Load(nil)
					require.NoError(t, err)
					assertConfig(t, config, map[string]interface{}{
						"string": "env_value",
					})
					os.Unsetenv("STRING")
				})
				t.Run("flag", func(t *testing.T) {
					flagSet := pflag.NewFlagSet("test", pflag.ExitOnError)
					flagSet.String("string", "test", "test flag")
					_ = flagSet.Parse([]string{"--string", "flag_value"})

					loader := NewLoader("./testdata/yaml")

					config, err := loader.Load(flagSet)
					require.NoError(t, err)
					assertConfig(t, config, map[string]interface{}{
						"string": "flag_value",
					})
				})
			})
		})

		t.Run("env", func(t *testing.T) {
			os.Setenv("STRING", "test")
			os.Setenv("INT", "10")
			os.Setenv("FLOAT", "10.5")
			os.Setenv("MAP", "one:one,two:two")
			os.Setenv("SLICE", "one,two,three")
			os.Setenv("DURATION", "10s")
			os.Setenv("PASSWORD", "abc")
			os.Setenv("CUSTOM", "two")

			config, err := Load(nil)
			require.NoError(t, err)
			assertConfig(t, config, nil)

			t.Run("default", func(t *testing.T) {
				os.Unsetenv("DURATION")

				RegisterDefaults(map[string]interface{}{
					"duration": 30 * time.Second,
				})

				config, err := Load(nil)
				require.NoError(t, err)
				assertConfig(t, config, map[string]interface{}{
					"duration": 30 * time.Second,
				})
			})

			t.Run("prefix", func(t *testing.T) {
				os.Clearenv()
				os.Setenv("PRE_STRING", "test")
				os.Setenv("PRE_INT", "10")
				os.Setenv("PRE_FLOAT", "10.5")
				os.Setenv("PRE_MAP", "one:one,two:two")
				os.Setenv("PRE_SLICE", "one,two,three")
				os.Setenv("PRE_DURATION", "10s")
				os.Setenv("PRE_PASSWORD", "abc")
				os.Setenv("PRE_CUSTOM", "two")

				loader := NewLoader()
				loader.SetEnvPrefix("pre")
				config, err := loader.Load(nil)
				require.NoError(t, err)
				assertConfig(t, config, nil)
			})

			os.Clearenv()
		})
	})

	t.Run("load into", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			loader := NewLoader("./testdata/json")
			err := loader.LoadInto(nil, nil)
			require.NoError(t, err)
		})

		t.Run("file", func(t *testing.T) {
			t.Run("json", func(t *testing.T) {
				loader := NewLoader("./testdata/json")

				var cfg TestConfig
				err := loader.LoadInto(nil, &cfg)
				require.NoError(t, err)
				assertTestConfig(t, cfg, nil)
			})

			t.Run("yaml", func(t *testing.T) {
				loader := NewLoader("./testdata/yaml")

				var cfg TestConfig
				err := loader.LoadInto(nil, &cfg)
				require.NoError(t, err)
				assertTestConfig(t, cfg, nil)
			})

			t.Run("override", func(t *testing.T) {
				t.Run("env", func(t *testing.T) {
					os.Setenv("STRING", "env_value")

					loader := NewLoader("./testdata/yaml")

					var config TestConfig
					err := loader.LoadInto(nil, &config)
					require.NoError(t, err)
					assertTestConfig(t, config, map[string]interface{}{
						"string": "env_value",
					})
					os.Unsetenv("STRING")
				})
				t.Run("flag", func(t *testing.T) {
					flagSet := pflag.NewFlagSet("test", pflag.ExitOnError)
					flagSet.String("string", "test", "test flag")
					_ = flagSet.Parse([]string{"--string", "flag_value"})

					loader := NewLoader("./testdata/yaml")

					var config TestConfig
					err := loader.LoadInto(flagSet, &config)
					require.NoError(t, err)
					assertTestConfig(t, config, map[string]interface{}{
						"string": "flag_value",
					})
				})
			})
		})

		t.Run("env", func(t *testing.T) {
			os.Setenv("STRING", "test")
			os.Setenv("INT", "10")
			os.Setenv("FLOAT", "10.5")
			os.Setenv("MAP", "one:one,two:two")
			os.Setenv("SLICE", "one,two,three")
			os.Setenv("DURATION", "10s")
			os.Setenv("CUSTOM", "two")

			var config TestConfig
			err := LoadInto(nil, &config)
			require.NoError(t, err)
			assertTestConfig(t, config, nil)

			t.Run("default", func(t *testing.T) {
				os.Unsetenv("DURATION")

				RegisterDefaults(map[string]interface{}{
					"duration": 30 * time.Second,
				})

				var config TestConfig
				err := LoadInto(nil, &config)
				require.NoError(t, err)
				assertTestConfig(t, config, map[string]interface{}{
					"duration": 30 * time.Second,
				})
			})

			t.Run("prefix", func(t *testing.T) {
				os.Clearenv()
				os.Setenv("PRE_STRING", "test")
				os.Setenv("PRE_INT", "10")
				os.Setenv("PRE_FLOAT", "10.5")
				os.Setenv("PRE_MAP", "one:one,two:two")
				os.Setenv("PRE_SLICE", "one,two,three")
				os.Setenv("PRE_DURATION", "10s")
				os.Setenv("PRE_CUSTOM", "two")

				var config TestConfig
				loader := NewLoader()
				loader.SetEnvPrefix("pre")
				err := loader.LoadInto(nil, &config)
				require.NoError(t, err)
				assertTestConfig(t, config, nil)
			})

			os.Clearenv()
		})
	})

	t.Run("loaded", func(t *testing.T) {
		loader := NewLoader("./testdata/json")
		cfg, err := loader.Load(nil)
		require.NoError(t, err)
		assert.Equal(t, cfg, loader.LoadedConfig())
	})

	t.Run("write", func(t *testing.T) {
		loader := NewLoader("./testdata/json")
		cfg, err := loader.Load(nil)
		require.NoError(t, err)

		b, err := json.Marshal(cfg.(*config).AllSettings())
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		n, err := cfg.WriteTo(buf)
		require.NoError(t, err)
		assert.Equal(t, n, int64(len(b)))
		assert.Equal(t, b, buf.Bytes())
	})

	t.Run("secure", func(t *testing.T) {
		os.Setenv("NESTED_FOO", "bar")
		expected := []byte(`{"custom":"two","duration":"10s","float":10.5,"int":10,"map":"**********","nested":{"foo":"**********"},"password":"**********","slice":"**********","string":"test"}`)

		var cfg TestSecureConfig
		loader := NewLoader("./testdata/json")
		err := loader.LoadInto(nil, &cfg)
		require.NoError(t, err)

		loaded := loader.LoadedConfig()

		buf := &bytes.Buffer{}
		n, err := loaded.WriteTo(buf)
		require.NoError(t, err)
		assert.Equal(t, n, int64(len(expected)))
		assert.Equal(t, expected, buf.Bytes())
		os.Clearenv()
	})

	t.Run("nested env", func(t *testing.T) {
		os.Setenv("STRING", "test")
		os.Setenv("NESTEDPTR_STRING", "test1")
		os.Setenv("NESTEDPTR_INT", "10")
		os.Setenv("NESTED_STRING", "test2")
		os.Setenv("NESTED_INT", "15")

		var config TestNestedConfig
		err := LoadInto(nil, &config)
		require.NoError(t, err)

		assert.Equal(t, "test", config.String)
		assert.Equal(t, "test2", config.Nested.String)
		assert.Equal(t, 15, config.Nested.Int)
		require.NotNil(t, config.NestedPtr)
		assert.Equal(t, "test1", config.NestedPtr.String)
		assert.Equal(t, 10, config.NestedPtr.Int)

		t.Run("prefix", func(t *testing.T) {
			os.Clearenv()
			os.Setenv("PRE_STRING", "pretest")
			os.Setenv("PRE_NESTEDPTR_STRING", "pretest1")
			os.Setenv("PRE_NESTEDPTR_INT", "100")
			os.Setenv("PRE_NESTED_STRING", "pretest2")
			os.Setenv("PRE_NESTED_INT", "150")

			var config TestNestedConfig
			loader := NewLoader()
			loader.SetEnvPrefix("pre")
			err := loader.LoadInto(nil, &config)
			require.NoError(t, err)

			assert.Equal(t, "pretest", config.String)
			assert.Equal(t, "pretest2", config.Nested.String)
			assert.Equal(t, 150, config.Nested.Int)
			require.NotNil(t, config.NestedPtr)
			assert.Equal(t, "pretest1", config.NestedPtr.String)
			assert.Equal(t, 100, config.NestedPtr.Int)
		})

		os.Clearenv()
	})
}

func expected(t *testing.T, overrides map[string]interface{}) map[string]interface{} {
	t.Helper()
	expected := map[string]interface{}{
		"string": "test",
		"int":    10,
		"float":  10.5,
		"map": map[string]string{
			"one": "one",
			"two": "two",
		},
		"slice": []string{
			"one",
			"two",
			"three",
		},
		"duration": 10 * time.Second,
		"password": "abc",
		"custom":   CustomTypeTwo,
	}
	for key, value := range overrides {
		expected[key] = value
	}
	return expected
}

func assertConfig(t *testing.T, config Config, overrides map[string]interface{}) {
	t.Helper()
	expected := expected(t, overrides)
	assert.Equal(t, expected["string"], config.GetString("string"))
	assert.Equal(t, expected["int"], config.GetInt("int"))
	assert.Equal(t, expected["float"], config.GetFloat64("float"))
	assert.Equal(t, expected["map"], config.GetStringMapString("map"))
	assert.Equal(t, expected["slice"], config.GetStringSlice("slice"))
	assert.Equal(t, expected["duration"], config.GetDuration("duration"))
	assert.Equal(t, expected["password"], config.GetString("password"))
	assert.Equal(t, "two", config.Get("custom")) // custom unmarshal doesn't run
}

// assertTestConfig asserts the config bound to the TestConfig struct matches
// what is expected
func assertTestConfig(t *testing.T, config TestConfig, overrides map[string]interface{}) {
	t.Helper()
	expected := expected(t, overrides)
	assert.Equal(t, expected["string"], config.String)
	assert.Equal(t, expected["int"], config.Int)
	assert.Equal(t, expected["float"], config.Float)
	assert.Equal(t, expected["map"], config.Map)
	assert.Equal(t, expected["slice"], config.Slice)
	assert.Equal(t, expected["duration"], config.Duration)
	assert.Equal(t, expected["custom"], config.Custom)
}
