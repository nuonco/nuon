package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func TestGetDeploymentOrderFromAppConfig(t *testing.T) {
	ctx := context.Background()

	component1 := app.Component{
		ID:   "comp1",
		Name: "Component 1",
		Type: "docker_build",
	}
	component2 := app.Component{
		ID:   "comp2",
		Name: "Component 2",
		Type: "container_build",
	}
	component3 := app.Component{
		ID:   "comp3",
		Name: "Component 3",
		Type: "helm_chart",
	}
	component4 := app.Component{
		ID:   "comp4",
		Name: "Component 4",
		Type: "terraform_module",
	}
	component5 := app.Component{
		ID:   "comp5",
		Name: "Component 5",
		Type: "helm_chart",
	}

	appConfigRooted := &app.AppConfig{
		ComponentConfigConnections: []app.ComponentConfigConnection{
			{
				ComponentID:            "comp1",
				Component:              component1,
				ComponentDependencyIDs: []string{},
			},
			{
				ComponentID:            "comp2",
				Component:              component2,
				ComponentDependencyIDs: []string{},
			},
			{
				ComponentID:            "comp3",
				Component:              component3,
				ComponentDependencyIDs: []string{component1.ID, component2.ID},
			},
			{
				ComponentID:            "comp4",
				Component:              component4,
				ComponentDependencyIDs: []string{component1.ID, component2.ID},
			},
			{
				ComponentID:            "comp5",
				Component:              component5,
				ComponentDependencyIDs: []string{component3.ID, component4.ID},
			},
		},
	}

	appConfigNonRooted := &app.AppConfig{
		ComponentConfigConnections: []app.ComponentConfigConnection{
			{
				ComponentID:            "comp1",
				Component:              component1,
				ComponentDependencyIDs: []string{},
			},
			{
				ComponentID:            "comp2",
				Component:              component2,
				ComponentDependencyIDs: []string{},
			},
			{
				ComponentID:            "comp3",
				Component:              component3,
				ComponentDependencyIDs: []string{component1.ID},
			},
		},
	}

	t.Run("valid deployment order with root and dependents", func(t *testing.T) {
		compIds := []string{"comp1", "comp2", "comp3"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.NoError(t, err)
		require.NotNil(t, order)

		expectedOrder := []string{"comp2", "comp1", "comp3"}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("valid deployment order with root", func(t *testing.T) {
		compIds := []string{"comp1", "comp2"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.NoError(t, err)
		require.NotNil(t, order)

		expectedOrder := []string{"comp2", "comp1"}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("valid deployment order with leaf", func(t *testing.T) {
		compIds := []string{"comp4", "comp5"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.NoError(t, err)
		require.NotNil(t, order)

		expectedOrder := []string{"comp4", "comp5"}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("partial component IDs", func(t *testing.T) {
		compIds := []string{"comp1", "comp3"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.NoError(t, err)
		require.NotNil(t, order)

		expectedOrder := []string{"comp1", "comp3"}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("no matching component IDs", func(t *testing.T) {
		compIds := []string{"comp9"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.Error(t, err)
		require.Nil(t, order)
	})

	t.Run("empty component IDs", func(t *testing.T) {
		compIds := []string{}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigRooted)
		require.NoError(t, err)
		require.NotNil(t, order)

		expectedOrder := []string{}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("unrooted graph", func(t *testing.T) {
		compIds := []string{"comp1", "comp2", "comp3"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, appConfigNonRooted)
		require.NoError(t, err)
		require.NotNil(t, order)
		expectedOrder := []string{"comp2", "comp1", "comp3"}
		assert.Equal(t, expectedOrder, order)
	})

	t.Run("cyclic dependency error", func(t *testing.T) {
		cyclicAppConfig := &app.AppConfig{
			ComponentConfigConnections: []app.ComponentConfigConnection{
				{
					ComponentID:            "comp1",
					Component:              component1,
					ComponentDependencyIDs: []string{"comp2"},
				},
				{
					ComponentID:            "comp2",
					Component:              component2,
					ComponentDependencyIDs: []string{"comp1"},
				},
			},
		}

		compIds := []string{"comp1", "comp2"}
		order, err := GetDeploymentOrderFromAppConfig(ctx, compIds, cyclicAppConfig)
		require.Error(t, err)
		assert.Nil(t, order)
	})
}
