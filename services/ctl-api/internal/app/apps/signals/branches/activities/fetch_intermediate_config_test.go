package activities

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchIntermediateConfigCapturesCustomerManagedSource(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0o755))
	files := map[string]string{
		"metadata.toml": `version = "v2"
[customer_managed]
runner_image_url = "registry.example.com/runner"
runner_image_tag = "v1"

[[customer_managed.platforms]]
target = "linux/amd64"
portal_binary_url = "https://artifacts.example.com/portal"
runner_binary_url = "https://artifacts.example.com/runner"
`,
		"sandbox.toml": `terraform_version = "1.11.3"
[public_repo]
repo = "nuonco/aws-eks-sandbox"
directory = "."
branch = "main"
`,
		"runner.toml": `runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`,
		"permissions.toml": `[provision_role]
name = "test-provision-role"
description = "test provision role"

[[provision_role.policies]]
name = "test-policy"
contents = "{}"

[maintenance_role]
name = "test-maintenance-role"
description = "test maintenance role"

[[maintenance_role.policies]]
name = "test-policy"
contents = "{}"

[deprovision_role]
name = "test-deprovision-role"
description = "test deprovision role"

[[deprovision_role.policies]]
name = "test-policy"
contents = "{}"
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644))
	}

	result, err := (&Activities{}).fetchIntermediateConfig(context.Background(), dir)
	require.NoError(t, err)
	require.NotNil(t, result.CustomerManaged)
	require.NotNil(t, result.SourceArchive)
	require.Equal(t, files["metadata.toml"], result.SourceArchive.Files["metadata.toml"])
	_, err = os.Stat(dir)
	require.ErrorIs(t, err, os.ErrNotExist)
}
