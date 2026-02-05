package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestInstallUnmarshal(t *testing.T) {
	testCases := []struct {
		name     string
		toml     string
		expected Install
	}{
		{
			name: "single input group",
			toml: `
name = 'test-install'

[[inputs]]
host = 'localhost'
port = '5432'
database = 'myapp'
`,
			expected: Install{
				Name: "test-install",
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"host":     "localhost",
							"port":     "5432",
							"database": "myapp",
						},
					},
				},
			},
		},
		{
			name: "empty input group",
			toml: `
name = 'minimal-install'

[[inputs]]
`,
			expected: Install{
				Name: "minimal-install",
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string(nil),
					},
				},
			},
		},
		{
			name: "input group with empty values",
			toml: `
name = 'install-with-empties'

[[inputs]]
required_field = 'value'
optional_field = ''
another_field = 'another_value'
`,
			expected: Install{
				Name: "install-with-empties",
				InputGroups: []InputGroup{
					{
						Group: "",
						Inputs: map[string]string{
							"required_field": "value",
							"optional_field": "",
							"another_field":  "another_value",
						},
					},
				},
			},
		},
		{
			name: "mixed groups",
			toml: `
name = 'mixed-install'

[[inputs]]
key1 = 'value1'

[[inputs]]
key2 = 'value2'

[[inputs]]
key3 = 'value3'
`,
			expected: Install{
				Name: "mixed-install",
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"key1": "value1",
						},
					},
					{
						Inputs: map[string]string{
							"key2": "value2",
						},
					},
					{
						Inputs: map[string]string{
							"key3": "value3",
						},
					},
				},
			},
		},
		{
			name: "install with aws account and inputs",
			toml: `
name = 'aws-install'
approval_option = 'approve-all'

[aws_account]
region = 'us-west-2'

[[inputs]]
app_name = 'my-app'
environment = 'staging'
`,
			expected: Install{
				Name:           "aws-install",
				ApprovalOption: InstallApprovalOptionApproveAll,
				AWSAccount: &AWSAccount{
					Region: "us-west-2",
				},
				InputGroups: []InputGroup{
					{
						Group: "",
						Inputs: map[string]string{
							"app_name":    "my-app",
							"environment": "staging",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result Install
			err := toml.Unmarshal([]byte(tc.toml), &result)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected.Name, result.Name)
			assert.Equal(t, tc.expected.ApprovalOption, result.ApprovalOption)

			if tc.expected.AWSAccount != nil {
				assert.NotNil(t, result.AWSAccount)
				assert.Equal(t, tc.expected.AWSAccount.Region, result.AWSAccount.Region)
			} else {
				assert.Nil(t, result.AWSAccount)
			}

			assert.Equal(t, len(tc.expected.InputGroups), len(result.InputGroups))
			for i, expectedGroup := range tc.expected.InputGroups {
				assert.Equal(t, expectedGroup.Group, result.InputGroups[i].Group,
					"InputGroup[%d].Group mismatch", i)
				assert.Equal(t, expectedGroup.Inputs, result.InputGroups[i].Inputs,
					"InputGroup[%d].Inputs mismatch", i)
			}
		})
	}
}

func TestInstallRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		install Install
	}{
		{
			name: "basic install with single input group",
			install: Install{
				Name: "roundtrip-test",
				InputGroups: []InputGroup{
					{
						Group: "database",
						Inputs: map[string]string{
							"host":     "localhost",
							"port":     "5432",
							"database": "testdb",
						},
					},
				},
			},
		},
		{
			name: "install with multiple input groups",
			install: Install{
				Name: "multi-group-test",
				InputGroups: []InputGroup{
					{
						Group: "database",
						Inputs: map[string]string{
							"host": "localhost",
							"port": "5432",
						},
					},
					{
						Group: "app-config",
						Inputs: map[string]string{
							"environment": "production",
							"debug":       "false",
						},
					},
				},
			},
		},
		{
			name: "install with aws account",
			install: Install{
				Name:           "aws-test",
				ApprovalOption: InstallApprovalOptionPrompt,
				AWSAccount: &AWSAccount{
					Region: "eu-west-1",
				},
				InputGroups: []InputGroup{
					{
						Group: "settings",
						Inputs: map[string]string{
							"key": "value",
						},
					},
				},
			},
		},
		{
			name: "install with empty input values",
			install: Install{
				Name: "empty-values-test",
				InputGroups: []InputGroup{
					{
						Group: "optional",
						Inputs: map[string]string{
							"required": "value",
							"optional": "",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal to TOML
			marshaled, err := toml.Marshal(tc.install)
			assert.Nil(t, err)

			// Unmarshal back
			var result Install
			err = toml.Unmarshal(marshaled, &result)
			assert.Nil(t, err)

			// Verify all fields match
			assert.Equal(t, tc.install.Name, result.Name)
			assert.Equal(t, tc.install.ApprovalOption, result.ApprovalOption)

			if tc.install.AWSAccount != nil {
				assert.NotNil(t, result.AWSAccount)
				assert.Equal(t, tc.install.AWSAccount.Region, result.AWSAccount.Region)
			}

			assert.Equal(t, len(tc.install.InputGroups), len(result.InputGroups))
			for i, expectedGroup := range tc.install.InputGroups {
				// Group field is preserved through comments but may not round-trip
				// through standard TOML marshal/unmarshal without custom handling
				assert.Equal(t, expectedGroup.Inputs, result.InputGroups[i].Inputs,
					"InputGroup[%d].Inputs mismatch", i)
			}
		})
	}
}

func TestInstallFileParseFromRawTOML(t *testing.T) {
	testCases := []struct {
		name            string
		rawTOML         string
		expectedInstall Install
	}{
		{
			name: "basic install with inputs",
			rawTOML: `name = "production-install"
approval_option = "approve-all"

[aws_account]
region = "us-east-1"

[[inputs]]
sub_domain = "whoami"
test = "required-value"
testdefault = "default-value"
`,
			expectedInstall: Install{
				Name:           "production-install",
				ApprovalOption: InstallApprovalOptionApproveAll,
				AWSAccount: &AWSAccount{
					Region: "us-east-1",
				},
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"sub_domain":  "whoami",
							"test":        "required-value",
							"testdefault": "default-value",
						},
					},
				},
			},
		},
		{
			name: "install with multiple input groups",
			rawTOML: `name = "staging-install"
approval_option = "prompt"

[aws_account]
region = "us-west-2"

[[inputs]]
environment = "staging"
debug = "true"
log_level = "debug"

[[inputs]]
redis_host = "redis.staging.example.com"
redis_port = "6379"
redis_db = "0"
`,
			expectedInstall: Install{
				Name:           "staging-install",
				ApprovalOption: InstallApprovalOptionPrompt,
				AWSAccount: &AWSAccount{
					Region: "us-west-2",
				},
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"environment": "staging",
							"debug":       "true",
							"log_level":   "debug",
						},
					},
					{
						Inputs: map[string]string{
							"redis_host": "redis.staging.example.com",
							"redis_port": "6379",
							"redis_db":   "0",
						},
					},
				},
			},
		},
		{
			name: "minimal install",
			rawTOML: `name = "minimal-install"

[[inputs]]
app_name = "my-app"
`,
			expectedInstall: Install{
				Name: "minimal-install",
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"app_name": "my-app",
						},
					},
				},
			},
		},
		{
			name: "install with empty and complex values",
			rawTOML: `name = "complex-install"
approval_option = "approve-all"

# group.name : help
[[inputs]]
required_field = "value"
optional_field = ""
url = "https://example.com/api"
port = "8080"
enabled = "true"
`,
			expectedInstall: Install{
				Name:           "complex-install",
				ApprovalOption: InstallApprovalOptionApproveAll,
				InputGroups: []InputGroup{
					{
						Inputs: map[string]string{
							"required_field": "value",
							"optional_field": "",
							"url":            "https://example.com/api",
							"port":           "8080",
							"enabled":        "true",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for test files
			tempDir := t.TempDir()
			installPath := filepath.Join(tempDir, "install.toml")

			// Write the raw TOML content to a file
			err := os.WriteFile(installPath, []byte(tc.rawTOML), 0o644)
			require.NoError(t, err, "failed to write install file")

			// Read the file back
			fileContent, err := os.ReadFile(installPath)
			require.NoError(t, err, "failed to read install file")

			// Parse the file content using TOML unmarshal
			var parsedInstall Install
			err = toml.Unmarshal(fileContent, &parsedInstall)
			require.NoError(t, err, "failed to parse install file")

			// Verify the parsed install matches expectations
			assert.Equal(t, tc.expectedInstall.Name, parsedInstall.Name, "install name mismatch")
			assert.Equal(t, tc.expectedInstall.ApprovalOption, parsedInstall.ApprovalOption,
				"approval option mismatch")

			// Check AWS account
			if tc.expectedInstall.AWSAccount != nil {
				require.NotNil(t, parsedInstall.AWSAccount, "AWS account should not be nil")
				assert.Equal(t, tc.expectedInstall.AWSAccount.Region, parsedInstall.AWSAccount.Region,
					"AWS region mismatch")
			} else {
				assert.Nil(t, parsedInstall.AWSAccount, "AWS account should be nil")
			}

			// Check input groups
			require.Equal(t, len(tc.expectedInstall.InputGroups), len(parsedInstall.InputGroups),
				"input group count mismatch")
			for i, expectedGroup := range tc.expectedInstall.InputGroups {
				assert.Equal(t, expectedGroup.Inputs, parsedInstall.InputGroups[i].Inputs,
					"InputGroup[%d].Inputs mismatch", i)
			}

			// Run Parse() to ensure it doesn't error
			err = parsedInstall.Parse()
			assert.NoError(t, err, "Parse() should not error")

			// Verify the file content matches what we wrote
			assert.Equal(t, tc.rawTOML, string(fileContent), "file content should match raw TOML")
		})
	}
}
