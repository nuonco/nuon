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
