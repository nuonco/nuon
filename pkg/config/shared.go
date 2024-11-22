package config

func ToTerraformVarsMap(inp []TerraformVariable) map[string]string {
	v := make(map[string]string, 0)
	for _, kv := range inp {
		v[kv.Name] = kv.Value
	}

	return v
}

func TerraformVariables(input []TerraformVariable) []string {
	vals := make([]string, 0)
	for _, inp := range input {
		vals = append(vals, inp.Value)
	}

	return vals
}

type TerraformVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}

type EnvironmentVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}
