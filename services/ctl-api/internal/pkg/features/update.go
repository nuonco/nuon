package features

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

	if err := syncControlPlaneBuildsAndOrgRunner(org.Features, updateFeatures); err != nil {
		return err
	}

	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Features: updateFeatures,
	})

	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}

	return nil
}

// syncControlPlaneBuildsAndOrgRunner keeps control-plane-builds and org-runner
// mutually exclusive: an org either builds on the control plane (no org runner)
// or has an org runner. Callers always send the full feature map, so we detect
// which of the two flags actually changed against the org's current state and
// derive the other from it. This lets disabling control-plane-builds re-enable
// org-runner, so a subsequent reprovision restores the org runner group.
func syncControlPlaneBuildsAndOrgRunner(current, updated map[string]bool) error {
	cpKey := string(app.OrgFeatureControlPlaneBuilds)
	runnerKey := string(app.OrgFeatureOrgRunner)

	newCP, hasCP := updated[cpKey]
	newRunner, hasRunner := updated[runnerKey]
	if !hasCP || !hasRunner {
		return nil
	}

	cpChanged := newCP != current[cpKey]
	runnerChanged := newRunner != current[runnerKey]

	switch {
	case cpChanged && !runnerChanged:
		updated[runnerKey] = !newCP
	case runnerChanged && !cpChanged:
		updated[cpKey] = !newRunner
	case cpChanged && runnerChanged && newCP == newRunner:
		return fmt.Errorf("control-plane-builds and org-runner are mutually exclusive and cannot both be set to %t", newCP)
	}

	return nil
}
