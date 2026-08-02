package kubernetescontexts

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync persists the app's named kubernetes_context bindings via the shared
// builder in internal/pkg/config/build, which the CreateAppKubernetesContexts
// handler also uses.
//
// This must run after components are synced: the source component must exist
// before its name can be resolved to an ID.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.KubernetesContexts == nil || len(cfg.KubernetesContexts.Contexts) == 0 {
		return nil
	}

	inputs := build.KubernetesContextInputsFromConfig(cfg.KubernetesContexts)

	names := make([]string, 0, len(inputs))
	for _, c := range inputs {
		names = append(names, c.ComponentName)
	}

	var components []app.Component
	if res := db.WithContext(ctx).
		Select("id", "name").
		Where(&app.Component{AppID: appID}).
		Where("name IN ?", names).
		Find(&components); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to look up source components for kubernetes contexts",
			Err:         res.Error,
		}
	}

	componentIDByName := make(map[string]string, len(components))
	for _, c := range components {
		componentIDByName[c.Name] = c.ID
	}

	obj, err := build.KubernetesContextsConfig(inputs, componentIDByName, appID, appConfigID)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-kubernetes-contexts",
			Description: err.Error(),
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app kubernetes contexts config",
			Err:         res.Error,
		}
	}

	return nil
}
