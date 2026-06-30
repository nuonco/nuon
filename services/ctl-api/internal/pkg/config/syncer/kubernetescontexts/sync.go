package kubernetescontexts

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Sync persists the app's named kubernetes_context bindings for this app config
// version. Each context resolves a peer component name to its ID so the binding
// has a stable FK, while the name is also persisted so it stays intelligible
// across app config versions if the component is renamed.
//
// This must run after components are synced: the source component must exist
// before its name can be resolved to an ID.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.KubernetesContexts == nil || len(cfg.KubernetesContexts.Contexts) == 0 {
		return nil
	}

	contexts := cfg.KubernetesContexts.Contexts

	names := make([]string, 0, len(contexts))
	for _, c := range contexts {
		names = append(names, c.Component)
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

	children := make([]app.AppKubernetesContextConfig, 0, len(contexts))
	for _, c := range contexts {
		compID, ok := componentIDByName[c.Component]
		if !ok {
			return sync.SyncErr{
				Resource:    "app-kubernetes-contexts",
				Description: fmt.Sprintf("kubernetes_context %q references unknown component %q", c.Name, c.Component),
			}
		}
		children = append(children, app.AppKubernetesContextConfig{
			AppID:               appID,
			AppConfigID:         appConfigID,
			Name:                c.Name,
			SourceComponentName: c.Component,
			SourceComponentID:   compID,
		})
	}

	obj := app.AppKubernetesContextsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Contexts:    children,
	}

	if res := db.WithContext(ctx).Create(&obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app kubernetes contexts config",
			Err:         res.Error,
		}
	}

	return nil
}
