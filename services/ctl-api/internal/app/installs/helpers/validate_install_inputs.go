package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) ValidateInstallInputs(ctx context.Context, appID string, inputs map[string]*string) error {
	var parentApp app.App
	res := s.db.WithContext(ctx).
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Preload("AppInputConfigs.AppInputs").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return fmt.Errorf("unable to find parent: %w", res.Error)
	}

	if len(parentApp.AppInputConfigs) < 1 {
		if len(inputs) > 0 {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid install inputs provided"),
				Description: "inputs provided on install, that are not defined on the app",
			}
		}

		return nil
	}

	// verify all of the inputs are set on the current sandbox config
	for _, inp := range parentApp.AppInputConfigs[0].AppInputs {
		if !inp.Required {
			continue
		}

		inputVal, ok := inputs[inp.Name]
		if !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s is a required input", inp.Name),
				Description: fmt.Sprintf("Please add a value value for the %s input", inp.Name),
			}
		}

		if inputVal == nil || len(*inputVal) < 1 {
			return stderr.ErrUser{
				Err:         fmt.Errorf("%s must be non-empty", inp.Name),
				Description: fmt.Sprintf("Please add a non-empty value for the %s input", inp.Name),
			}
		}
	}

	return nil
}
