package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
)

func TestInstallMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		install  Install
		expected string
	}{
		{
			name: "basic install with multiple input groups",
			install: Install{
				InputGroups: []InputGroup{
					{
						Group: "test group",
						Inputs: map[string]string{
							"one": "onessss",
							"two": "tworrss",
						},
					},
					{
						Group: "test group 2",
						Inputs: map[string]string{
							"onesssssss": "one",
							"twoosss":    "two",
						},
					},
				},
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = ''

# input.group : test group
[[inputs]]
one = 'onessss'
two = 'tworrss'

# input.group : test group 2
[[inputs]]
onesssssss = 'one'
twoosss = 'two'
`,
		},
		{
			name:    "empty install",
			install: Install{},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = ''
`,
		},
		{
			name: "install with name",
			install: Install{
				Name: "my-install",
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = 'my-install'
`,
		},
		{
			name: "single input group",
			install: Install{
				InputGroups: []InputGroup{
					{
						Group: "database",
						Inputs: map[string]string{
							"host":     "localhost",
							"port":     "5432",
							"database": "myapp",
						},
					},
				},
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = ''

# input.group : database
[[inputs]]
database = 'myapp'
host = 'localhost'
port = '5432'
`,
		},
		{
			name: "input group with empty inputs",
			install: Install{
				InputGroups: []InputGroup{
					{
						Group:  "empty-group",
						Inputs: map[string]string{},
					},
				},
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = ''

# input.group : empty-group
[[inputs]]
`,
		},
		{
			name: "input group with empty values",
			install: Install{
				InputGroups: []InputGroup{
					{
						Group: "optional-configs",
						Inputs: map[string]string{
							"required_field": "value",
							"optional_field": "",
							"another_field":  "another_value",
						},
					},
				},
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = ''

# input.group : optional-configs
[[inputs]]
another_field = 'another_value'
optional_field = ''
required_field = 'value'
`,
		},
		{
			name: "install with name and input groups",
			install: Install{
				Name: "production-install",
				InputGroups: []InputGroup{
					{
						Group: "app-config",
						Inputs: map[string]string{
							"environment": "production",
							"debug":       "false",
						},
					},
				},
			},
			expected: `# #:schema https://api.nuon.co/v1/general/config-schema?type=install
name = 'production-install'

# input.group : app-config
[[inputs]]
debug = 'false'
environment = 'production'
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := toml.Marshal(tc.install)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, string(b))
		})
	}
}
