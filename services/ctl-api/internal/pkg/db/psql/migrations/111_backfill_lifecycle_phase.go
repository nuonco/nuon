package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration111BackfillLifecyclePhase(ctx context.Context, db *gorm.DB) error {
	var colExists bool
	db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'installs' AND column_name = 'lifecycle_status'
		)
	`).Scan(&colExists)

	if !colExists {
		return nil
	}

	if res := db.WithContext(ctx).Exec(`
		UPDATE installs
		SET lifecycle_phase = jsonb_build_object(
			'phase', lifecycle_status->>'status',
			'description', lifecycle_status->>'status_human_description',
			'updated_at', COALESCE((lifecycle_status->>'created_at_ts')::bigint, 0),
			'updated_by', COALESCE(lifecycle_status->>'created_by_id', ''),
			'metadata', COALESCE(lifecycle_status->'metadata', '{}'::jsonb),
			'history', COALESCE(
				(SELECT jsonb_agg(jsonb_build_object(
					'phase', h->>'status',
					'description', h->>'status_human_description',
					'updated_at', COALESCE((h->>'created_at_ts')::bigint, 0),
					'updated_by', COALESCE(h->>'created_by_id', ''),
					'metadata', COALESCE(h->'metadata', '{}'::jsonb)
				)) FROM jsonb_array_elements(lifecycle_status->'history') h),
				'[]'::jsonb
			)
		)
		WHERE lifecycle_status IS NOT NULL
		  AND lifecycle_phase IS NULL;
	`); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).Exec(`
		ALTER TABLE installs DROP COLUMN IF EXISTS lifecycle_status;
	`); res.Error != nil {
		return res.Error
	}

	return nil
}
