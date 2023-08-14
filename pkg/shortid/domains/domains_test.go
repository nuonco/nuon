package domains

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

type testFakeObj struct {
	AppID         string `faker:"appID"`
	ArtifactID    string `faker:"artifactID"`
	AWSSettingsID string `faker:"awsSettingsID"`
	BuildID       string `faker:"buildID"`
	CanaryID      string `faker:"canaryID"`
	ComponentID   string `faker:"componentID"`
	DeployID      string `faker:"deployID"`
	DeploymentID  string `faker:"deploymentID"`
	InstallID     string `faker:"installID"`
	InstanceID    string `faker:"instanceID"`
	OrgID         string `faker:"orgID"`
	SandboxID     string `faker:"sandboxID"`
	SecretID      string `faker:"secretID"`
	UserID        string `faker:"userID"`
}

func TestFakerStructTags(t *testing.T) {
	var obj testFakeObj
	err := faker.FakeData(&obj)
	assert.NoError(t, err)
	assert.Len(t, obj.AppID, 26)
	assert.Equal(t, "app", obj.AppID[:3])
	assert.Len(t, obj.ArtifactID, 26)
	assert.Equal(t, "art", obj.ArtifactID[:3])
	assert.Len(t, obj.AWSSettingsID, 26)
	assert.Equal(t, "aws", obj.AWSSettingsID[:3])
	assert.Len(t, obj.BuildID, 26)
	assert.Equal(t, "bld", obj.BuildID[:3])
	assert.Len(t, obj.CanaryID, 26)
	assert.Equal(t, "can", obj.CanaryID[:3])
	assert.Len(t, obj.ComponentID, 26)
	assert.Equal(t, "cmp", obj.ComponentID[:3])
	assert.Len(t, obj.DeploymentID, 26)
	assert.Equal(t, "dpl", obj.DeploymentID[:3])
	assert.Len(t, obj.InstallID, 26)
	assert.Equal(t, "inl", obj.InstallID[:3])
	assert.Len(t, obj.InstanceID, 26)
	assert.Equal(t, "ins", obj.InstanceID[:3])
	assert.Len(t, obj.OrgID, 26)
	assert.Equal(t, "org", obj.OrgID[:3])
	assert.Len(t, obj.SandboxID, 26)
	assert.Equal(t, "snb", obj.SandboxID[:3])
	assert.Len(t, obj.SecretID, 26)
	assert.Equal(t, "sec", obj.SecretID[:3])
	assert.Len(t, obj.UserID, 26)
	assert.Equal(t, "usr", obj.UserID[:3])
}

func TestNewAppID(t *testing.T) {
	t.Run("get valid ID for App", func(t *testing.T) {
		id := NewAppID()
		assert.Len(t, id, 26)
		assert.Equal(t, "app", id[:3])
	})
}

func TestNewArtifactID(t *testing.T) {
	t.Run("get valid ID for Artifact", func(t *testing.T) {
		id := NewArtifactID()
		assert.Len(t, id, 26)
		assert.Equal(t, "art", id[:3])
	})
}

func TestNewAWSSettingsID(t *testing.T) {
	t.Run("get valid ID for AWSSettings", func(t *testing.T) {
		id := NewAWSAccountID()
		assert.Len(t, id, 26)
		assert.Equal(t, "aws", id[:3])
	})
}

func TestNewBuildID(t *testing.T) {
	t.Run("get valid ID for Build", func(t *testing.T) {
		id := NewBuildID()
		assert.Len(t, id, 26)
		assert.Equal(t, "bld", id[:3])
	})
}

func TestNewCanaryID(t *testing.T) {
	t.Run("get valid ID for Canary", func(t *testing.T) {
		id := NewCanaryID()
		assert.Len(t, id, 26)
		assert.Equal(t, "can", id[:3])
	})
}

func TestNewDeploymentID(t *testing.T) {
	t.Run("get valid ID for Deployment", func(t *testing.T) {
		id := NewDeploymentID()
		assert.Len(t, id, 26)
		assert.Equal(t, "dpl", id[:3])
	})
}

func TestNewDomainID(t *testing.T) {
	t.Run("get valid ID for Domain", func(t *testing.T) {
		id := NewDomainID()
		assert.Len(t, id, 26)
		assert.Equal(t, "dom", id[:3])
	})
}

func TestNewInstallID(t *testing.T) {
	t.Run("get valid ID for Install", func(t *testing.T) {
		id := NewInstallID()
		assert.Len(t, id, 26)
		assert.Equal(t, "inl", id[:3])
	})
}

func TestNewInstanceID(t *testing.T) {
	t.Run("get valid ID for Instance", func(t *testing.T) {
		id := NewInstanceID()
		assert.Len(t, id, 26)
		assert.Equal(t, "ins", id[:3])
	})
}

func TestNewOrgID(t *testing.T) {
	t.Run("get valid ID for Org", func(t *testing.T) {
		id := NewOrgID()
		assert.Len(t, id, 26)
		assert.Equal(t, "org", id[:3])
	})
}

func TestNewSandboxID(t *testing.T) {
	t.Run("get valid ID for Sandbox", func(t *testing.T) {
		id := NewSandboxID()
		assert.Len(t, id, 26)
		assert.Equal(t, "snb", id[:3])
	})
}

func TestNewSecretID(t *testing.T) {
	t.Run("get valid ID for Secret", func(t *testing.T) {
		id := NewSecretID()
		assert.Len(t, id, 26)
		assert.Equal(t, "sec", id[:3])
	})
}

func TestNewUserID(t *testing.T) {
	t.Run("get valid ID for User", func(t *testing.T) {
		id := NewUserID()
		assert.Len(t, id, 26)
		assert.Equal(t, "usr", id[:3])
	})
}
