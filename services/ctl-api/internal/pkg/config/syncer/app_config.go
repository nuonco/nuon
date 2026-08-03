package syncer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// fetchState gives orphan detection a baseline. Missing or unparseable state
// leaves it empty rather than failing the sync.
func (s *syncer) fetchState(ctx context.Context) {
	var prev app.AppConfig
	res := s.db.WithContext(ctx).
		Select("id", "state", "created_at").
		Where(app.AppConfig{AppID: s.appID}).
		Where("id != ?", s.appConfigID).
		Where("state != ''").
		Order("created_at DESC").
		First(&prev)
	if res.Error != nil {
		return
	}

	var prevState sync.State
	if err := json.Unmarshal([]byte(prev.State), &prevState); err != nil {
		return
	}

	s.prevState = &prevState
}

// persistState writes state the CLI sync reads back on its next run.
func (s *syncer) persistState(ctx context.Context) error {
	if orphans := s.orphanedResult(); orphans != nil {
		if s.state.Result == nil {
			s.state.Result = &sync.Result{}
		}
		s.state.Result.OrphanedComponents = orphans.OrphanedComponents
		s.state.Result.OrphanedActions = orphans.OrphanedActions
		s.state.Result.OrphanedRunbooks = orphans.OrphanedRunbooks
	}

	stateJSON, err := json.Marshal(s.state)
	if err != nil {
		return sync.SyncInternalErr{
			Description: "unable to serialize sync state",
			Err:         err,
		}
	}

	res := s.db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where("id = ?", s.appConfigID).
		Update("state", string(stateJSON))
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to persist sync state",
			Err:         fmt.Errorf("unable to persist sync state: %w", res.Error),
		}
	}

	return nil
}
