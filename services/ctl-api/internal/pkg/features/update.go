package features

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Features) Enable(ctx context.Context, orgID string, features map[string]bool) error {
	if err := s.validateOrgFeatures(features); err != nil {
		return errors.Wrap(err, "unable to validate org features")
	}

	if err := s.updateOrgFeatures(ctx, orgID, features); err != nil {
		return errors.Wrap(err, "unable to validate org features")
	}

	return nil
}

func (s *Features) validateOrgFeatures(features map[string]bool) error {
	orgFeatures := make(map[string]bool)
	if _, ok := features["all"]; ok {
		return nil
	}

	for _, value := range app.GetFeatures() {
		orgFeatures[string(value)] = true
	}
	for feature := range features {
		if _, ok := orgFeatures[feature]; !ok {
			return fmt.Errorf("invalid feature: %s", feature)
		}
	}

	return nil
}

func (s *Features) updateOrgFeatures(ctx context.Context, orgID string, updateFeatures map[string]bool) error {
	var org app.Org
	if res := s.db.WithContext(ctx).First(&org, "id = ?", orgID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get org")
	}

	if allValue, ok := updateFeatures["all"]; ok {
		for feature := range org.Features {
			updateFeatures[feature] = allValue
		}
	} else {
		// add features from org.Features not in features
		for feature, enabled := range org.Features {
			if _, ok := updateFeatures[feature]; !ok {
				updateFeatures[feature] = enabled
			}
		}
	}

	// Remove the "all" key from updateFeatures if it exists
	delete(updateFeatures, "all")

	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Features: updateFeatures,
	})

	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}

	return nil
}
