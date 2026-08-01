package inputs

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app input configuration via the shared builders in
// internal/pkg/config/build, which the CreateAppInputConfig handler also uses.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID, orgID string, state *sync.State) error {
	groups, inputs := build.InputsFromConfig(cfg)

	inputCfg := build.InputConfig(groups, appID, appConfigID, orgID)
	if res := db.WithContext(ctx).Create(inputCfg); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app input config",
			Err:         fmt.Errorf("unable to create app input groups: %w", res.Error),
		}
	}

	objs, err := build.AppInputs(inputs, inputCfg)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-inputs",
			Description: err.Error(),
		}
	}

	if len(objs) > 0 {
		if res := db.WithContext(ctx).Create(&objs); res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create app inputs",
				Err:         fmt.Errorf("unable to create app inputs: %w", res.Error),
			}
		}
	}

	state.InputConfigID = inputCfg.ID
	return nil
}
