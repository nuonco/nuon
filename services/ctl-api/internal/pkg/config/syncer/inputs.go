package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// syncAppInput creates the app input configuration with groups and inputs.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_input_config.go
func (s *syncer) syncAppInput(ctx context.Context) error {
	// Handle nil inputs config
	if s.cfg.Inputs == nil {
		cfg := app.AppInputConfig{
			AppConfigID:    s.appConfigID,
			OrgID:          s.orgID,
			AppID:          s.appID,
			AppInputGroups: []app.AppInputGroup{},
		}

		res := s.db.WithContext(ctx).Create(&cfg)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create empty app input config",
				Err:         fmt.Errorf("unable to create app input config: %w", res.Error),
			}
		}

		s.state.InputConfigID = cfg.ID
		return nil
	}

	// Validate required inputs with existing installs
	if err := s.validateRequiredInputsWithInstalls(ctx); err != nil {
		return sync.SyncErr{
			Resource:    "app-inputs",
			Description: "validation failed",
			Err:         err,
		}
	}

	// Create groups
	groups := make([]app.AppInputGroup, 0, len(s.cfg.Inputs.Groups))
	for idx, group := range s.cfg.Inputs.Groups {
		groups = append(groups, app.AppInputGroup{
			Name:        group.Name,
			Description: group.Description,
			DisplayName: group.DisplayName,
			Index:       idx,
		})
	}

	cfg := app.AppInputConfig{
		AppConfigID:    s.appConfigID,
		OrgID:          s.orgID,
		AppID:          s.appID,
		AppInputGroups: groups,
	}

	res := s.db.WithContext(ctx).Create(&cfg)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app input config",
			Err:         fmt.Errorf("unable to create app input groups: %w", res.Error),
		}
	}

	// Create inputs
	if len(s.cfg.Inputs.Inputs) > 0 {
		inputs := make([]app.AppInput, 0, len(s.cfg.Inputs.Inputs))

		for idx, input := range s.cfg.Inputs.Inputs {
			// Find group ID
			var groupID string
			for _, group := range cfg.AppInputGroups {
				if group.Name == input.Group {
					groupID = group.ID
					break
				}
			}

			// Determine source
			source := app.AppInputSourceVendor
			if input.UserConfigurable {
				source = app.AppInputSourceCustomer
			}

			// Validate JSON defaults
			if input.Type == "json" && input.Default != nil {
				defaultStr := fmt.Sprintf("%v", input.Default)
				if defaultStr != "" && !json.Valid([]byte(defaultStr)) {
					return sync.SyncErr{
						Resource:    "app-inputs",
						Description: fmt.Sprintf("input %s has invalid JSON default value", input.Name),
					}
				}
			}

			var defaultVal string
			if input.Default != nil {
				defaultVal = fmt.Sprintf("%v", input.Default)
			}

			inputType := generics.ValOrDefault(input.Type, "string")

			inputs = append(inputs, app.AppInput{
				OrgID:            cfg.OrgID,
				AppInputConfigID: cfg.ID,
				AppInputGroupID:  groupID,
				Name:             input.Name,
				Description:      input.Description,
				DisplayName:      input.DisplayName,
				Required:         input.Required,
				Default:          defaultVal,
				Sensitive:        input.Sensitive,
				Type:             app.AppInputType(inputType),
				Internal:         input.Internal,
				Index:            idx,
				Source:           source,
			})
		}

		res := s.db.WithContext(ctx).Create(&inputs)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to create app inputs",
				Err:         fmt.Errorf("unable to create app inputs: %w", res.Error),
			}
		}
	}

	s.state.InputConfigID = cfg.ID
	return nil
}

// validateRequiredInputsWithInstalls ensures new required inputs have default values when installs exist.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_input_config.go
func (s *syncer) validateRequiredInputsWithInstalls(ctx context.Context) error {
	var installCount int64
	res := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("app_id = ?", s.appID).
		Count(&installCount)

	if res.Error != nil {
		return fmt.Errorf("unable to check install count: %w", res.Error)
	}

	if installCount == 0 {
		return nil
	}

	// Get existing input names
	var existingInputConfig app.AppInputConfig
	res = s.db.WithContext(ctx).
		Where("app_id = ?", s.appID).
		Preload("AppInputs").
		Order("created_at DESC").
		First(&existingInputConfig)

	existingInputNames := make(map[string]bool)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("unable to check existing input config: %w", res.Error)
	}

	for _, inp := range existingInputConfig.AppInputs {
		existingInputNames[inp.Name] = true
	}

	// Validate NEW required inputs have defaults
	for _, input := range s.cfg.Inputs.Inputs {
		if existingInputNames[input.Name] {
			continue
		}

		if !input.Required {
			continue
		}

		if input.Default == nil || fmt.Sprintf("%v", input.Default) == "" {
			return stderr.ErrUser{
				Err: fmt.Errorf("required input '%s' is missing a default value", input.Name),
				Description: fmt.Sprintf(
					"Cannot add new required input '%s' without a default value because %d install(s) exist for this app. "+
						"When existing installs are present, all new required inputs must have default values to ensure "+
						"data integrity during migration. Existing inputs from previous syncs are not affected. "+
						"Please add a default value to this input in your app config.",
					input.Name, installCount,
				),
			}
		}
	}

	return nil
}
