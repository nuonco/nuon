package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// maxStackVersionRuns bounds how many runs each stack version carries in a
// response. A version gains a run every time the customer applies their stack,
// so a frequently reapplied version would otherwise drag the whole version list
// with it.
const maxStackVersionRuns = 10

// rankedStackVersionRunIDs keeps the newest runs per version rather than the
// newest overall, which a plain LIMIT would give.
const rankedStackVersionRunIDs = `
SELECT id FROM (
	SELECT id, row_number() OVER (
		PARTITION BY install_stack_version_id ORDER BY created_at DESC
	) AS rn
	FROM install_stack_version_runs
	WHERE install_stack_version_id IN ? AND deleted_at = 0
) ranked
WHERE rn <= ?`

// attachStackVersionRuns loads the newest runs for each of the given stack
// versions and hangs them off the versions in place.
//
// This replaces a `Preload("Runs", ... Limit(n))`: GORM applies a preload's
// limit to the whole result set rather than per parent, so the versions that
// sort first take the entire budget and every older version comes back with no
// runs at all — and so with no outputs to show for its own phone home.
func (s *service) attachStackVersionRuns(ctx context.Context, versions []app.InstallStackVersion, perVersion int) error {
	if len(versions) == 0 {
		return nil
	}

	versionIDs := make([]string, 0, len(versions))
	for i := range versions {
		versionIDs = append(versionIDs, versions[i].ID)
	}

	var runIDs []string
	if err := s.db.WithContext(ctx).
		Raw(rankedStackVersionRunIDs, versionIDs, perVersion).
		Scan(&runIDs).Error; err != nil {
		return fmt.Errorf("unable to rank install stack version runs: %w", err)
	}
	if len(runIDs) == 0 {
		return nil
	}

	var runs []app.InstallStackVersionRun
	if err := s.db.WithContext(ctx).
		Where("id IN ?", runIDs).
		Order("created_at DESC").
		Find(&runs).Error; err != nil {
		return fmt.Errorf("unable to get install stack version runs: %w", err)
	}

	byVersion := make(map[string][]app.InstallStackVersionRun, len(versionIDs))
	for _, run := range runs {
		byVersion[run.InstallStackVersionID] = append(byVersion[run.InstallStackVersionID], run)
	}

	for i := range versions {
		versions[i].Runs = byVersion[versions[i].ID]
	}

	return nil
}
