package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestNotApplicableNeverMasksAssessedResources(t *testing.T) {
	at := time.Now().Truncate(time.Second)
	row := func(kind, name, health string) app.InstallComponentResourceState {
		return app.InstallComponentResourceState{
			InstallComponentID: "inc1",
			Provider:           providerKubernetes,
			Kind:               kind,
			Name:               name,
			Health:             health,
			ObservedAt:         at,
		}
	}

	t.Run("not-applicable listed first does not mask healthy", func(t *testing.T) {
		rows := []app.InstallComponentResourceState{
			row("ClusterRole", "rbac", "not-applicable"),
			row("ConfigMap", "cm", "not-applicable"),
			row("Deployment", "api", "healthy"),
		}
		got := collapseComponentHealthRows(rows)["inc1"]
		assert.Len(t, got, 1)
		assert.Equal(t, app.InstallComponentHealthStatusHealthy, got[0].Health)
	})

	t.Run("a genuinely bad resource still wins", func(t *testing.T) {
		rows := []app.InstallComponentResourceState{
			row("ClusterRole", "rbac", "not-applicable"),
			row("Deployment", "api", "healthy"),
			row("Deployment", "admin", "degraded"),
		}
		got := collapseComponentHealthRows(rows)["inc1"]
		assert.Equal(t, app.InstallComponentHealthStatusDegraded, got[0].Health)
	})

	t.Run("unknown outranks not-applicable when nothing is assessed", func(t *testing.T) {
		rows := []app.InstallComponentResourceState{
			row("ClusterRole", "rbac", "not-applicable"),
			row("Certificate", "cert", "unknown"),
		}
		got := collapseComponentHealthRows(rows)["inc1"]
		assert.Equal(t, app.InstallComponentHealthStatusUnknown, got[0].Health)
	})

	t.Run("only non-assessable resources is still not-applicable", func(t *testing.T) {
		rows := []app.InstallComponentResourceState{
			row("ClusterRole", "rbac", "not-applicable"),
			row("ConfigMap", "cm", "not-applicable"),
		}
		got := collapseComponentHealthRows(rows)["inc1"]
		assert.Equal(t, app.InstallComponentHealthStatusNotApplicable, got[0].Health)
	})
}
