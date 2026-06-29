package features

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Features) ToggleForAllOrgs(ctx context.Context, features map[string]bool) error {
	if _, ok := features["all"]; ok {
		return fmt.Errorf("'all' is not supported when updating features across all orgs")
	}

	if err := s.validateOrgFeatures(features); err != nil {
		return errors.Wrap(err, "unable to validate org features")
	}

	patch, err := json.Marshal(features)
	if err != nil {
		return errors.Wrap(err, "unable to marshal features patch")
	}

	res := s.db.WithContext(ctx).
		Model(&app.Org{}).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Update("features", gorm.Expr("COALESCE(features, '{}'::jsonb) || ?::jsonb", string(patch)))
	if res.Error != nil {
		return fmt.Errorf("unable to update org features: %w", res.Error)
	}

	return nil
}
