package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const maxStackVersionRuns = 10

const rankedStackVersionRunIDs = `
SELECT id FROM (
	SELECT id, row_number() OVER (
		PARTITION BY install_stack_version_id ORDER BY created_at DESC
	) AS rn
	FROM install_stack_version_runs
	WHERE install_stack_version_id IN ? AND deleted_at = 0
) ranked
WHERE rn <= ?`

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
