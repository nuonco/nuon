package components

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
)

// EnsureComponent creates a component if it doesn't exist, using the shared helpers
// for full initialization (queue creation, dependencies, install components).
func EnsureComponent(ctx context.Context, db *gorm.DB, helpers *componenthelpers.Helpers, comp *config.Component, appID string) error {
	_, err := getComponent(ctx, db, comp.Name, appID)
	if err == nil {
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if component %s exists", comp.Name),
			Err:         err,
		}
	}

	_, err = helpers.CreateComponent(ctx, &componenthelpers.CreateComponentParams{
		AppID:        appID,
		Name:         comp.Name,
		VarName:      comp.VarName,
		Dependencies: comp.Dependencies,
	})
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create component %s", comp.Name),
			Err:         err,
		}
	}

	return nil
}

// SyncComponent updates a component and creates its configuration based on type.
func SyncComponent(ctx context.Context, db *gorm.DB, comp *config.Component, appID, appConfigID string, state *sync.State) error {
	apiComp, err := getComponent(ctx, db, comp.Name, appID)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get component %s", comp.Name),
			Err:         err,
		}
	}

	updates := app.Component{
		Name:    comp.Name,
		VarName: comp.VarName,
		Type:    app.ComponentType(comp.Type.APIType()),
	}

	res := db.WithContext(ctx).
		Model(apiComp).
		Updates(updates)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update component %s", comp.Name),
			Err:         res.Error,
		}
	}

	// Create component config based on type
	var configID string
	var checksum string

	switch comp.Type.APIType() {
	case "docker_build":
		configID, checksum, err = SyncDockerBuildComponent(ctx, db, comp, apiComp.ID, appID, appConfigID)
		if err != nil {
			return err
		}
	case "helm_chart":
		configID, checksum, err = SyncHelmComponent(ctx, db, comp, apiComp.ID, appID, appConfigID)
		if err != nil {
			return err
		}
	case "terraform_module":
		configID, checksum, err = SyncTerraformModuleComponent(ctx, db, comp, apiComp.ID, appID, appConfigID)
		if err != nil {
			return err
		}
	case "external_image":
		// TODO: Implement external image component sync
		configID = ""
		checksum = ""
	case "job":
		// TODO: Implement job component sync
		configID = ""
		checksum = ""
	case "kubernetes_manifest":
		configID, checksum, err = SyncKubernetesManifestComponent(ctx, db, comp, apiComp.ID, appID, appConfigID)
		if err != nil {
			return err
		}
	default:
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unsupported component type: %s", comp.Type.APIType()),
			Err:         fmt.Errorf("component type %s is not supported", comp.Type.APIType()),
		}
	}

	state.Components = append(state.Components, sync.ComponentState{
		Name:     apiComp.Name,
		Type:     comp.Type.APIType(),
		ID:       apiComp.ID,
		ConfigID: configID,
		Checksum: checksum,
	})

	return nil
}

// getComponent finds a component by name.
func getComponent(ctx context.Context, db *gorm.DB, name string, appID string) (*app.Component, error) {
	var comp app.Component
	res := db.WithContext(ctx).
		Where("app_id = ? AND name = ?", appID, name).
		First(&comp)

	if res.Error != nil {
		return nil, res.Error
	}

	return &comp, nil
}
