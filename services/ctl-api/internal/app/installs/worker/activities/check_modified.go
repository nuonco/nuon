package activities

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type CheckModifiedRequest struct {
	InstallID   string
	PartialName string
	LastKnownAt time.Time
}

type CheckModifiedResponse struct {
	Changed          bool
	LatestModifiedAt time.Time
}

// CheckModified checks if a partial has been modified since the last known time.
// Each partial maps to a cheap SELECT MAX(updated_at) query.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 10s
func (a *Activities) CheckModified(ctx context.Context, req *CheckModifiedRequest) (*CheckModifiedResponse, error) {
	var latestAt time.Time
	var err error

	switch req.PartialName {
	case "org":
		err = a.db.WithContext(ctx).
			Raw("SELECT o.updated_at FROM orgs o JOIN installs i ON i.org_id = o.id WHERE i.id = ? AND i.deleted_at = 0", req.InstallID).
			Scan(&latestAt).Error
	case "app":
		err = a.db.WithContext(ctx).
			Raw(`SELECT GREATEST(
				(SELECT a.updated_at FROM apps a JOIN installs i ON i.app_id = a.id WHERE i.id = ? AND i.deleted_at = 0),
				COALESCE((SELECT MAX(s.updated_at) FROM app_secrets s JOIN installs i ON i.app_id = s.app_id WHERE i.id = ? AND i.deleted_at = 0), '1970-01-01'::timestamptz)
			)`, req.InstallID, req.InstallID).
			Scan(&latestAt).Error
	case "domain", "sandbox":
		err = a.db.WithContext(ctx).
			Raw("SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz) FROM install_sandbox_runs WHERE install_id = ?", req.InstallID).
			Scan(&latestAt).Error
	case "runner":
		err = a.db.WithContext(ctx).
			Raw(`SELECT r.updated_at FROM runners r
				JOIN runner_groups rg ON r.runner_group_id = rg.id
				WHERE rg.owner_id = ? AND rg.owner_type = 'installs'
				ORDER BY r.created_at DESC LIMIT 1`, req.InstallID).
			Scan(&latestAt).Error
	case "cloud":
		err = a.db.WithContext(ctx).
			Raw("SELECT updated_at FROM installs WHERE id = ? AND deleted_at = 0", req.InstallID).
			Scan(&latestAt).Error
	case "actions":
		err = a.db.WithContext(ctx).
			Raw("SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz) FROM install_action_workflows WHERE install_id = ?", req.InstallID).
			Scan(&latestAt).Error
	case "inputs":
		err = a.db.WithContext(ctx).
			Raw("SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz) FROM install_inputs WHERE install_id = ?", req.InstallID).
			Scan(&latestAt).Error
	case "components":
		err = a.db.WithContext(ctx).
			Raw(`SELECT GREATEST(
				COALESCE((SELECT MAX(updated_at) FROM install_components WHERE install_id = ?), '1970-01-01'::timestamptz),
				COALESCE((SELECT MAX(id.updated_at) FROM install_deploys id JOIN install_components ic ON id.install_component_id = ic.id WHERE ic.install_id = ?), '1970-01-01'::timestamptz)
			)`, req.InstallID, req.InstallID).
			Scan(&latestAt).Error
	case "stack":
		err = a.db.WithContext(ctx).
			Raw(`SELECT COALESCE(MAX(isv.created_at), '1970-01-01'::timestamptz)
				FROM install_stack_versions isv
				JOIN install_stacks ist ON isv.install_stack_id = ist.id
				WHERE ist.install_id = ?`, req.InstallID).
			Scan(&latestAt).Error
	case "secrets":
		err = a.db.WithContext(ctx).
			Raw(`SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamptz)
				FROM runner_jobs
				WHERE install_id = ? AND type = 'sync-secrets'`, req.InstallID).
			Scan(&latestAt).Error
	default:
		return &CheckModifiedResponse{Changed: true, LatestModifiedAt: time.Now()}, nil
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &CheckModifiedResponse{Changed: false}, nil
		}
		return nil, err
	}

	return &CheckModifiedResponse{
		Changed:          latestAt.After(req.LastKnownAt),
		LatestModifiedAt: latestAt,
	}, nil
}
