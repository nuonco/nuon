package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type Resources struct {
	kind  permissions.ResourceKind
	orgID string
	query *gorm.DB
}

type targetRow struct {
	ID            string
	OrgID         string
	AppID         string
	InstallID     string
	RelatedJobIDs []string `gorm:"serializer:json"`
}

func NewResources(targets *gorm.DB, kind permissions.ResourceKind, orgID string) (*Resources, error) {
	if err := (permissions.Scope{Kind: kind, OrgID: orgID, ID: orgID}).ValidateShape(); err != nil {
		return nil, err
	}
	query := targets.Session(&gorm.Session{NewDB: true}).Table("(?) AS permission_target", targets).
		Where(&targetRow{OrgID: orgID})
	return &Resources{kind: kind, orgID: orgID, query: query.Session(&gorm.Session{})}, nil
}

func (r *Resources) Resolve(ctx context.Context, id string) (permissions.Target, error) {
	if err := (permissions.Scope{Kind: r.kind, OrgID: r.orgID, ID: id}).ValidateShape(); err != nil {
		return permissions.Target{}, err
	}
	var row targetRow
	result := r.query.WithContext(ctx).Where(&targetRow{ID: id}).Scan(&row)
	if result.Error != nil {
		return permissions.Target{}, fmt.Errorf("resolve permission target: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	target := permissions.Target{ResolvedTarget: permissions.ResolvedTarget{
		Scope: permissions.Scope{Kind: r.kind, OrgID: row.OrgID, ID: row.ID}, AppID: row.AppID, InstallID: row.InstallID,
	}}
	for _, jobID := range row.RelatedJobIDs {
		target.RelatedJobs = append(target.RelatedJobs, permissions.Scope{Kind: permissions.KindRunnerJob, OrgID: row.OrgID, ID: jobID})
	}
	if err := target.ValidateShape(); err != nil {
		return permissions.Target{}, err
	}
	return target, nil
}
