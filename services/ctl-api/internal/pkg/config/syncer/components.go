package syncer

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// getComponent finds a component by name.
func (s *syncer) getComponent(ctx context.Context, name string) (*app.Component, error) {
	var comp app.Component
	res := s.db.WithContext(ctx).
		Where("app_id = ? AND name = ?", s.appID, name).
		First(&comp)

	if res.Error != nil {
		return nil, res.Error
	}

	return &comp, nil
}

// ensureComponent creates a component if it doesn't exist.
func (s *syncer) ensureComponent(ctx context.Context, comp *config.Component) error {
	_, err := s.getComponent(ctx, comp.Name)
	if err == nil {
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if component %s exists", comp.Name),
			Err:         err,
		}
	}

	newComp := app.Component{
		AppID: s.appID,
		Name:  comp.Name,
	}

	res := s.db.WithContext(ctx).Create(&newComp)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create component %s", comp.Name),
			Err:         res.Error,
		}
	}

	return nil
}

// syncComponent updates a component and creates its configuration.
func (s *syncer) syncComponent(ctx context.Context, comp *config.Component) error {
	apiComp, err := s.getComponent(ctx, comp.Name)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get component %s", comp.Name),
			Err:         err,
		}
	}

	updates := app.Component{
		Name:         comp.Name,
		VarName:      comp.VarName,
		Dependencies: comp.Dependencies,
	}

	res := s.db.WithContext(ctx).
		Model(apiComp).
		Updates(updates)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update component %s", comp.Name),
			Err:         res.Error,
		}
	}

	// TODO: Create component config based on type
	configID := ""
	checksum := ""

	s.state.Components = append(s.state.Components, sync.ComponentState{
		Name:     apiComp.Name,
		Type:     comp.Type.APIType(),
		ID:       apiComp.ID,
		ConfigID: configID,
		Checksum: checksum,
	})

	return nil
}
