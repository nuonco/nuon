package config

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
)

func TestInstallValidate_CloudFormation(t *testing.T) {
	tests := map[string]struct {
		config  string
		wantErr string
	}{
		"no overrides preserves legacy installs": {},
		"legacy stack overrides do not require a cloud connection": {
			config: "[stack_overrides]\nvpc_nested_template_url = 'https://example.com/vpc.yaml'",
		},
		"overrides require a cloud connection": {
			config:  "[cloudformation.root.parameters]\nRunnerInstanceType = 't3.small'",
			wantErr: "cloudformation requires an AWS cloud_connection",
		},
		"empty overrides still require a connection": {
			config:  "[cloudformation]",
			wantErr: "cloudformation requires an AWS cloud_connection",
		},
		"literal templated and empty values are accepted": {
			config: `cloud_connection = "acme-aws"
[cloudformation.root.parameters]
RunnerInstanceType = "t3.small"
InstallDomain = "{{ .nuon.inputs.inputs.domain }}"
OptionalValue = ""
[cloudformation.nested_stacks.database.parameters]
InstanceClass = "db.r7g.large"
BackupRetentionDays = "{{ .nuon.inputs.inputs.retention }}"
OptionalValue = ""`,
		},
		"explicit clear is accepted": {
			config: "cloud_connection = 'acme-aws'\n[cloudformation.root.parameters]\n[cloudformation.nested_stacks]",
		},
		"parameter names cannot be templated": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.root.parameters]\n'{{ .nuon.id }}' = 'value'",
			wantErr: "must not be templated",
		},
		"invalid CloudFormation parameter names are rejected": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.nested_stacks.database.parameters]\n'Instance-Class' = 'value'",
			wantErr: "contain only letters and digits",
		},
		"unavailable state is rejected at validation": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.root.parameters]\nDomain = '{{ .nuon.components.api.outputs.domain }}'",
			wantErr: "not populated when the install stack is generated",
		},
		"malformed value templates are rejected": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.root.parameters]\nInstallDomain = '{{ .nuon.inputs.inputs.domain'",
			wantErr: "not a valid template",
		},
		"nested parameter templates are validated": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.nested_stacks.database.parameters]\nInstanceClass = '{{ .nuon.inputs.inputs.instance_class'",
			wantErr: "not a valid template",
		},
		"nested stack names cannot be templated": {
			config:  "cloud_connection = 'acme-aws'\n[cloudformation.nested_stacks.'{{ .nuon.id }}'.parameters]\nInstanceClass = 'db.r7g.large'",
			wantErr: "must be a nonempty literal",
		},
		"GCP cannot use AWS parameters": {
			config:  "cloud_connection = 'acme-gcp'\n[gcp_account]\n[cloudformation.root.parameters]\nRunnerInstanceType = 't3.small'",
			wantErr: "requires an AWS cloud_connection",
		},
		"Azure cannot use AWS parameters": {
			config:  "cloud_connection = 'acme-azure'\n[azure_account]\n[cloudformation.root.parameters]\nRunnerInstanceType = 't3.small'",
			wantErr: "requires an AWS cloud_connection",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var install Install
			require.NoError(t, toml.Unmarshal([]byte(tt.config), &install))
			err := install.Validate()
			if tt.wantErr != "" {
				var configErr ErrConfig
				require.ErrorAs(t, err, &configErr)
				assert.Contains(t, configErr.Description, tt.wantErr)
				return
			}
			require.NoError(t, err)
			encoded, err := toml.Marshal(install)
			require.NoError(t, err)
			var roundtrip Install
			require.NoError(t, toml.Unmarshal(encoded, &roundtrip))
			assert.Equal(t, install.CloudFormation, roundtrip.CloudFormation)
		})
	}
}

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
			expected: `# install
name = ''

# input.group: test group
[[inputs]]
one = 'onessss'
two = 'tworrss'

# input.group: test group 2
[[inputs]]
onesssssss = 'one'
twoosss = 'two'
`,
		},
		{
			name:    "empty install",
			install: Install{},
			expected: `# install
name = ''
`,
		},
		{
			name: "install with name",
			install: Install{
				Name: "my-install",
			},
			expected: `# install
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
			expected: `# install
name = ''

# input.group: database
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
			expected: `# install
name = ''

# input.group: empty-group
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
			expected: `# install
name = ''

# input.group: optional-configs
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
			expected: `# install
name = 'production-install'

# input.group: app-config
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

# group.name: help
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

func TestInstallComponentOverrides_ParseAndFlatten(t *testing.T) {
	rawTOML := `name = "prod"

[components.vpc]
tf_vars = """
cidr = "10.0.0.0/16"
"""

[components.clickhouse]
helm_values = """
replicas: 3
"""
`

	var install Install
	require.NoError(t, toml.Unmarshal([]byte(rawTOML), &install))
	require.NoError(t, install.Parse())

	require.Len(t, install.Components, 2)
	assert.Contains(t, install.Components["vpc"].TFVars, "cidr")
	assert.Contains(t, install.Components["clickhouse"].HelmValues, "replicas")

	flat := install.FlattenedInputs()
	assert.Equal(t, install.Components["vpc"].TFVars, flat[TFVarsOverrideInputName("vpc")])
	assert.Equal(t, install.Components["clickhouse"].HelmValues, flat[HelmValuesOverrideInputName("clickhouse")])
}

func TestInstallValidate_ComponentOverrideSyntax(t *testing.T) {
	t.Run("valid overrides pass", func(t *testing.T) {
		install := Install{
			Components: map[string]ComponentOverride{
				"vpc":        {TFVars: "cidr = \"10.0.0.0/16\"\n"},
				"clickhouse": {HelmValues: "replicas: 3\n"},
			},
		}
		require.NoError(t, install.Validate())
	})

	t.Run("invalid helm yaml is rejected", func(t *testing.T) {
		install := Install{
			Components: map[string]ComponentOverride{
				"clickhouse": {HelmValues: "replicas: [3\n"},
			},
		}
		err := install.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "helm_values")
	})

	t.Run("invalid tf_vars is rejected", func(t *testing.T) {
		install := Install{
			Components: map[string]ComponentOverride{
				"vpc": {TFVars: "cidr = =\n"},
			},
		}
		err := install.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tf_vars")
	})

	t.Run("empty overrides pass", func(t *testing.T) {
		install := Install{
			Components: map[string]ComponentOverride{
				"vpc": {},
			},
		}
		require.NoError(t, install.Validate())
	})
}

func TestInstallComponentOverrides_EmptyNotFlattened(t *testing.T) {
	install := Install{
		Components: map[string]ComponentOverride{
			"vpc": {}, // both fields empty
		},
	}
	flat := install.FlattenedInputs()
	_, hasTF := flat[TFVarsOverrideInputName("vpc")]
	_, hasHelm := flat[HelmValuesOverrideInputName("vpc")]
	assert.False(t, hasTF, "empty tf_vars should not be flattened")
	assert.False(t, hasHelm, "empty helm_values should not be flattened")
}

func TestInstallDiff_OverrideRemovalDetected(t *testing.T) {
	upstream := &Install{
		AWSAccount: &AWSAccount{},
		Components: map[string]ComponentOverride{
			"vpc":   {TFVars: "cidr = \"10.0.0.0/16\"\n"},
			"redis": {HelmValues: "replicas: 3\n"},
		},
	}

	// Local config drops the vpc override entirely but keeps redis.
	local := &Install{
		AWSAccount: &AWSAccount{},
		Components: map[string]ComponentOverride{
			"redis": {HelmValues: "replicas: 3\n"},
		},
	}

	d, err := local.Diff(upstream)
	require.NoError(t, err)

	summary := d.Summary()
	assert.True(t, summary.HasChanged, "removing an override must be detected as a change")
	assert.Equal(t, 1, summary.Removed, "the cleared vpc tf_vars should count as one removal")
}

func TestInstallDiff_OverrideRemovalIgnoresNormalInputs(t *testing.T) {
	upstream := &Install{
		AWSAccount: &AWSAccount{},
		InputGroups: []InputGroup{
			{Inputs: map[string]string{"domain": "nuon.run"}},
		},
	}

	// Local omits the normal "domain" input; this must NOT be treated as a
	// removal (normal inputs are additive).
	local := &Install{
		AWSAccount:  &AWSAccount{},
		InputGroups: []InputGroup{},
	}

	d, err := local.Diff(upstream)
	require.NoError(t, err)

	assert.Equal(t, 0, d.Summary().Removed, "omitting a normal input must not be a removal")
}

// A config that declares a target account must not diff against an upstream that
// reports the same one. If GenerateCLIInstallConfig ever stops emitting these
// fields, `nuon apps sync` reports drift on every run and this test fails.
func TestInstallDiff_TargetIdentifiersDoNotDrift(t *testing.T) {
	for _, tt := range []struct {
		name  string
		local *Install
	}{
		{
			name:  "aws account id",
			local: &Install{AWSAccount: &AWSAccount{Region: "us-west-2", AccountID: "123456789012"}},
		},
		{
			name:  "azure subscription id",
			local: &Install{AzureAccount: &AzureAccount{Location: "eastus", SubscriptionID: "sub-a"}},
		},
		{
			name:  "gcp project id",
			local: &Install{GCPAccount: &GCPAccount{Region: "us-central1", ProjectID: "proj-a"}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, err := tt.local.Diff(tt.local)
			require.NoError(t, err)
			assert.False(t, d.Summary().HasChanged,
				"an identical config must not report drift")
		})
	}
}

func TestInstallDiff_TargetIdentifierChangeIsDetected(t *testing.T) {
	upstream := &Install{AWSAccount: &AWSAccount{Region: "us-west-2", AccountID: "123456789012"}}
	local := &Install{AWSAccount: &AWSAccount{Region: "us-west-2", AccountID: "999999999999"}}

	d, err := local.Diff(upstream)
	require.NoError(t, err)
	assert.True(t, d.Summary().HasChanged, "a changed account id must be visible in the diff")
}

// The AWS branch used to dereference upstreamInstall.AWSAccount without a nil check,
// unlike its GCP and Azure siblings.
func TestInstallDiff_NilUpstreamAWSAccountDoesNotPanic(t *testing.T) {
	local := &Install{AWSAccount: &AWSAccount{Region: "us-west-2", AccountID: "123456789012"}}
	upstream := &Install{Name: "no-aws-account-block"}

	d, err := local.Diff(upstream)
	require.NoError(t, err)
	assert.True(t, d.Summary().HasChanged)
}

func TestInstallTargetIdentifiersRoundTripThroughTOML(t *testing.T) {
	raw := `
name = "my-install"

[aws_account]
region = "us-west-2"
account_id = "123456789012"
`
	var install Install
	require.NoError(t, toml.Unmarshal([]byte(raw), &install))
	require.NotNil(t, install.AWSAccount)
	assert.Equal(t, "us-west-2", install.AWSAccount.Region)
	assert.Equal(t, "123456789012", install.AWSAccount.AccountID)
}

func TestInstallValidate_Labels(t *testing.T) {
	t.Run("static labels pass", func(t *testing.T) {
		install := Install{
			Name:   "prod",
			Labels: map[string]string{"env": "production", "team": "platform"},
		}
		require.NoError(t, install.Validate())
	})

	t.Run("valid dynamic label values pass", func(t *testing.T) {
		install := Install{
			Name: "prod",
			Labels: map[string]string{
				"region":  "{{ .nuon.cloud_account.aws.region }}",
				"version": "{{ .nuon.components.api.outputs.version }}",
			},
		}
		require.NoError(t, install.Validate())
	})

	t.Run("unparsable template is rejected", func(t *testing.T) {
		install := Install{
			Name:   "prod",
			Labels: map[string]string{"region": "{{ .nuon.cloud_account.aws.region "},
		}
		err := install.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})

	t.Run("templated key is rejected", func(t *testing.T) {
		install := Install{
			Name:   "prod",
			Labels: map[string]string{"{{ .nuon.id }}": "value"},
		}
		err := install.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "label key")
	})

	t.Run("braces without nuon namespace stay literal", func(t *testing.T) {
		install := Install{
			Name:   "prod",
			Labels: map[string]string{"note": "{{ literal braces }}"},
		}
		require.NoError(t, install.Validate())
	})
}

func TestInstallValidate_RunnerImageOverrides(t *testing.T) {
	enabled := true

	tests := map[string]struct {
		install Install
		wantErr bool
	}{
		"auto-update only": {
			install: Install{Name: "prod", StackOverrides: &InstallStackOverrides{RunnerImageAutoUpdate: &enabled}},
		},
		"valid ami pin": {
			install: Install{Name: "prod", StackOverrides: &InstallStackOverrides{RunnerImageID: "ami-0abc123"}},
		},
		"malformed ami pin": {
			install: Install{Name: "prod", StackOverrides: &InstallStackOverrides{RunnerImageID: "ami-bogus!"}},
			wantErr: true,
		},
		"non-ami pin on a gcp install": {
			install: Install{
				Name:           "prod",
				GCPAccount:     &GCPAccount{ProjectID: "acme", Region: "us-central1"},
				StackOverrides: &InstallStackOverrides{RunnerImageID: "projects/acme/global/images/runner"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.install.Validate()

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestInstallStackOverrides_HasOverrides_RunnerImage(t *testing.T) {
	disabled := false

	tests := map[string]struct {
		overrides *InstallStackOverrides
		want      bool
	}{
		"nil":                  {overrides: nil, want: false},
		"empty":                {overrides: &InstallStackOverrides{}, want: false},
		"auto-update disabled": {overrides: &InstallStackOverrides{RunnerImageAutoUpdate: &disabled}, want: true},
		"image pin":            {overrides: &InstallStackOverrides{RunnerImageID: "ami-0abc123"}, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.overrides.HasOverrides())
		})
	}
}

func TestInstallDiff_PatchProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		value := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,12}`, true))
		for _, hasInput := range []bool{false, true} {
			for _, hasOverride := range []bool{false, true} {
				for _, enabled := range []bool{false, true} {
					upstream := &Install{
						InputGroups: []InputGroup{{Inputs: map[string]string{"region": value}}},
						Components:  map[string]ComponentOverride{"api": {TFVars: "region = " + strconv.Quote(value)}},
						Telemetry:   &InstallTelemetry{Enabled: boolPtr(enabled)},
					}
					local := &Install{}
					if hasInput {
						local.InputGroups = []InputGroup{{Inputs: map[string]string{"region": value}}}
					}
					if hasOverride {
						local.Components = upstream.Components
					}
					before, err := json.Marshal(upstream)
					require.NoError(t, err)
					got, err := local.Diff(upstream)
					require.NoError(t, err)
					require.Equal(t, !hasOverride, got.Summary().HasChanged, "omitted inputs and telemetry preserve state; omitted overrides clear")
					require.Equal(t, map[bool]int{true: 0, false: 1}[hasOverride], got.Summary().Removed)
					local.Telemetry = &InstallTelemetry{Enabled: boolPtr(!enabled)}
					got, err = local.Diff(upstream)
					require.NoError(t, err)
					require.Equal(t, 1, got.Summary().Changed, "explicit telemetry false must be observable")
					after, err := json.Marshal(upstream)
					require.NoError(t, err)
					require.Equal(t, before, after)
				}
			}
		}
	})
}

func TestInstallDiff_MapPresenceAndDeterminismProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		labels := hegel.Draw(t, hegel.Maps(hegel.FromRegex(`[a-z]{1,8}`, true), hegel.Text()).MinSize(2).MaxSize(8))
		inputs := hegel.Draw(t, hegel.Maps(hegel.FromRegex(`[a-z]{1,8}`, true), hegel.Text()).MinSize(2).MaxSize(8))
		old := &Install{Labels: maps.Clone(labels), InputGroups: []InputGroup{{Inputs: maps.Clone(inputs)}}}
		local := &Install{Labels: labels, InputGroups: []InputGroup{{Inputs: inputs}}}
		first, err := local.Diff(old)
		require.NoError(t, err)
		require.False(t, first.Summary().HasChanged)
		for range 3 {
			next, err := local.Diff(old)
			require.NoError(t, err)
			require.Equal(t, first, next, "map insertion history cannot change machine output")
		}
		for _, pair := range [][2]*Install{
			{{}, {Labels: map[string]string{"empty": ""}}},
			{{Labels: map[string]string{"empty": ""}}, {}},
			{{}, {InputGroups: []InputGroup{{Inputs: map[string]string{"empty": ""}}}}},
		} {
			got, err := pair[1].Diff(pair[0])
			require.NoError(t, err)
			require.True(t, got.Summary().HasChanged, "an empty value is present, not absent")
		}
	})
}

func TestInstallDiff_OptionalBooleanProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		states := []*bool{nil, boolPtr(false), boolPtr(true)}
		for _, oldState := range states {
			for _, desired := range states {
				makeInstall := func(state *bool) *Install {
					result := &Install{Name: "acme"}
					if state != nil {
						result.Telemetry = &InstallTelemetry{Enabled: state}
						result.ComponentToggles = map[string]bool{name: *state}
					}
					return result
				}
				old, next := makeInstall(oldState), makeInstall(desired)
				result, err := next.Diff(old)
				require.NoError(t, err)
				telemetryChanged := desired != nil && (oldState == nil || *desired != *oldState)
				telemetry := findChild(result, "telemetry.enabled")
				if desired == nil {
					require.Nil(t, telemetry, "omitted telemetry preserves upstream setting")
				} else {
					require.Equal(t, telemetryChanged, telemetry.DirectSummary().HasChanged)
					require.Contains(t, telemetry.Diff.Diff, "'"+strconv.FormatBool(*desired)+"'")
				}
				toggle := findChild(findChild(result, "inputs"), EnabledOverrideInputName(name))
				toggleChanged := (oldState == nil) != (desired == nil) || oldState != nil && desired != nil && *oldState != *desired
				if oldState == nil && desired == nil {
					require.Nil(t, toggle)
				} else {
					require.Equal(t, toggleChanged, toggle.DirectSummary().HasChanged, "false is explicit, omission clears the override")
				}
				require.Equal(t, telemetryChanged || toggleChanged, result.DirectSummary().HasChanged)
				encoded, err := json.Marshal(next)
				require.NoError(t, err)
				var decoded Install
				require.NoError(t, json.Unmarshal(encoded, &decoded))
				roundTrip, err := decoded.Diff(old)
				require.NoError(t, err)
				require.Equal(t, result, roundTrip, "transport must retain explicit false")
			}
		}
		for _, pair := range [][2]*OperationRolesConfig{{nil, {Type: OperationRuleConfigTypeMatrix}}, {{Type: OperationRuleConfigTypeMatrix}, nil}} {
			require.False(t, (&AppConfig{OperationRoles: pair[1]}).Diff(&AppConfig{OperationRoles: pair[0]}).Summary().HasChanged, "absent operation rules use the parser's empty matrix default")
		}
	})
}
