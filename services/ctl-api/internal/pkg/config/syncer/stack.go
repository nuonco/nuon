package syncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// syncAppCloudFormationStack creates the app CloudFormation stack configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_stack_config.go
func (s *syncer) syncAppCloudFormationStack(ctx context.Context) error {
	if s.cfg.Stack == nil {
		return nil
	}

	appCloudFormationStackConfig := app.AppStackConfig{
		Type:                    app.StackType(s.cfg.Stack.Type),
		AppConfigID:             s.appConfigID,
		AppID:                   s.appID,
		Name:                    s.cfg.Stack.Name,
		Description:             s.cfg.Stack.Description,
		VPCNestedTemplateURL:    s.cfg.Stack.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: s.cfg.Stack.RunnerNestedTemplateURL,
	}

	res := s.db.WithContext(ctx).Create(&appCloudFormationStackConfig)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app stack config",
			Err:         res.Error,
		}
	}

	return nil
}
