package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) GetStateID(id string, hasher hash.Hash) string {
	hasher.Write([]byte(id))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (s *Helpers) GetLockID(rawData []byte) (string, error) {
	lock := &app.TerraformLock{}
	err := json.Unmarshal(rawData, &lock)
	if err != nil {
		return "", err
	}

	return lock.ID, nil
}

func (s *Helpers) GetTerraformState(ctx context.Context, stateID string) (*app.TerraformState, error) {
	tfState := &app.TerraformState{}

	res := s.db.WithContext(ctx).
		Where("state_id = ?", stateID).
		Order("revision DESC").
		Limit(1).
		Find(tfState)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get terraform state: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		// No record found for the given stateID
		return nil, nil
	}

	return tfState, nil
}

func (s *Helpers) InsertTerraformState(ctx context.Context, sid string, tfState *app.TerraformState) error {
	if tfState == nil {
		return fmt.Errorf("terraform state is nil")
	}

	var latestRevision int
	err := s.db.WithContext(ctx).
		Model(&app.TerraformState{}).
		Where("state_id = ?", sid).
		Select("COALESCE(MAX(revision), 0)").
		Scan(&latestRevision).Error
	if err != nil {
		return fmt.Errorf("failed to fetch latest revision: %w", err)
	}

	tfState.ID = sid
	tfState.Revision = latestRevision + 1

	res := s.db.WithContext(ctx).Create(tfState)
	if res.Error != nil {
		return fmt.Errorf("failed to insert new terraform state: %w", res.Error)
	}

	return nil
}
