package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		id := NewAWSSettingsID()
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
