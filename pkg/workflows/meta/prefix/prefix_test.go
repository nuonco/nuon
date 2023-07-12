package prefix

import (
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestInstancePath(t *testing.T) {
	obj := generics.GetFakeObj[instance]()

	prefix := InstancePath(obj.OrgID, obj.AppID, obj.ComponentID, obj.DeploymentID, obj.InstallID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"deployment", obj.DeploymentID},
		{"component", obj.ComponentID},
		{"install", obj.InstallID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
	assert.NotContains(t, prefix, "phase=")
}

func TestInstanceOutputPath(t *testing.T) {
	obj := generics.GetFakeObj[instance]()

	prefix := InstanceOutputPath(obj.OrgID, obj.AppID, obj.ComponentID, obj.InstallID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"component", obj.ComponentID},
		{"install", obj.InstallID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
	assert.NotContains(t, prefix, "deployment=")
	assert.NotContains(t, prefix, "phase=")
}

func TestInstancePhasePath(t *testing.T) {
	obj := generics.GetFakeObj[instance]()

	prefix := InstancePhasePath(obj.OrgID, obj.AppID, obj.ComponentID, obj.DeploymentID, obj.InstallID, obj.Phase)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"deployment", obj.DeploymentID},
		{"component", obj.ComponentID},
		{"install", obj.InstallID},
		{"phase", obj.Phase},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestInstallPath(t *testing.T) {
	obj := generics.GetFakeObj[install]()

	prefix := InstallPath(obj.OrgID, obj.AppID, obj.InstallID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"install", obj.InstallID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestDeploymentPhasePath(t *testing.T) {
	obj := generics.GetFakeObj[deployment]()

	prefix := DeploymentPhasePath(obj.OrgID, obj.AppID, obj.ComponentName, obj.DeploymentID, obj.Phase)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"deployment", obj.DeploymentID},
		{"component", obj.ComponentName},
		{"phase", obj.Phase},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestDeploymentPath(t *testing.T) {
	obj := generics.GetFakeObj[deployment]()

	prefix := DeploymentPath(obj.OrgID, obj.AppID, obj.ComponentName, obj.DeploymentID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"deployment", obj.DeploymentID},
		{"component", obj.ComponentName},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
	assert.NotContains(t, prefix, "phase=")
}

func TestBuildPath(t *testing.T) {
	obj := generics.GetFakeObj[build]()

	prefix := BuildPath(obj.OrgID, obj.AppID, obj.ComponentID, obj.BuildID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"component", obj.ComponentID},
		{"build", obj.BuildID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestAppPath(t *testing.T) {
	obj := generics.GetFakeObj[app]()

	prefix := AppPath(obj.OrgID, obj.AppID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestOrgPath(t *testing.T) {
	obj := generics.GetFakeObj[org]()

	prefix := OrgPath(obj.OrgID)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestOrgComponentPath(t *testing.T) {
	obj := generics.GetFakeObj[orgComponent]()

	prefix := OrgComponentPath(obj.OrgID, obj.ComponentName)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"component", obj.ComponentName},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}

func TestSandboxPath(t *testing.T) {
	obj := generics.GetFakeObj[sandbox]()

	prefix := SandboxPath(obj.OrgID, obj.AppID, obj.InstallID, obj.SandboxName, obj.SandboxVersion)
	expectedKVs := [][2]string{
		{"org", obj.OrgID},
		{"app", obj.AppID},
		{"install", obj.InstallID},
		{"sandbox", obj.SandboxName},
		{"version", obj.SandboxVersion},
	}
	for _, kv := range expectedKVs {
		assert.Contains(t, prefix, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}
}
