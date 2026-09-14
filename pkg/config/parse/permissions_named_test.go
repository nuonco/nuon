package parse

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestParseDirNamedIAMPolicies(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "permissions", "policies"), 0o755))

	files := map[string]string{
		"metadata.toml": `version = "v2"
display_name = "named policy parse"
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
		"permissions/provision.toml": `type = "provision"
name = "provision"
description = "provision"
[[named_policies]]
name = "install-{{.nuon.install.id}}-logs"
[[policies]]
name = "inline"
contents = """{"Version":"2012-10-17","Statement":[]}"""
`,
		"permissions/maintenance.toml": `type = "maintenance"
name = "maintenance"
description = "maintenance"
[[policies]]
name = "inline"
contents = """{"Version":"2012-10-17","Statement":[]}"""
`,
		"permissions/deprovision.toml": `type = "deprovision"
name = "deprovision"
description = "deprovision"
[[policies]]
name = "inline"
contents = """{"Version":"2012-10-17","Statement":[]}"""
`,
		"permissions/policies/logs.toml": `name = "install-{{.nuon.install.id}}-logs"
description = "shared logs"
contents = """{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"logs:*","Resource":"*"}]}"""
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644))
	}

	result, err := ParseDirWithSource(context.Background(), ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.NotNil(t, result.Config.Permissions)
	require.Len(t, result.Config.Permissions.NamedPolicies, 1)
	require.Equal(t, "install-{{.nuon.install.id}}-logs", result.Config.Permissions.NamedPolicies[0].Name)
	require.Contains(t, result.Config.Permissions.NamedPolicies[0].Contents, "logs:*")
	require.Equal(t, []string{"install-{{.nuon.install.id}}-logs"}, config.NamedPolicyRefNames(result.Config.Permissions.ProvisionRole.NamedPolicies))
	require.Equal(t, "permissions/policies/logs.toml", result.Source.Members["permission_policy:logs"])
	require.Equal(t, "permissions/provision.toml", result.Source.Members["permission:provision"])
}
