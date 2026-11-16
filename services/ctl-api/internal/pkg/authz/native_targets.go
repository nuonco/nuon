package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type NativeResources struct {
	kind  permissions.ResourceKind
	orgID string
	query *gorm.DB
}

type nativeTargetRow struct {
	ID            string
	OrgID         string
	AppID         string
	InstallID     string
	RelatedJobIDs []string `gorm:"serializer:json"`
}

func NewNativeResources(targets *gorm.DB, kind permissions.ResourceKind, orgID string) (*NativeResources, error) {
	if err := (permissions.Scope{Kind: kind, OrgID: orgID, ID: orgID}).ValidateShape(); err != nil {
		return nil, err
	}
	query := targets.Session(&gorm.Session{NewDB: true}).Table("(?) AS native_target", targets).
		Where(&nativeTargetRow{OrgID: orgID})
	return &NativeResources{kind: kind, orgID: orgID, query: query.Session(&gorm.Session{})}, nil
}

func (r *NativeResources) Resolve(ctx context.Context, id string) (permissions.NativeTarget, error) {
	if err := (permissions.Scope{Kind: r.kind, OrgID: r.orgID, ID: id}).ValidateShape(); err != nil {
		return permissions.NativeTarget{}, err
	}
	var row nativeTargetRow
	result := r.query.WithContext(ctx).Where(&nativeTargetRow{ID: id}).Scan(&row)
	if result.Error != nil {
		return permissions.NativeTarget{}, fmt.Errorf("resolve native target: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return permissions.NativeTarget{}, gorm.ErrRecordNotFound
	}
	target := permissions.NativeTarget{ResolvedTarget: permissions.ResolvedTarget{
		Scope: permissions.Scope{Kind: r.kind, OrgID: row.OrgID, ID: row.ID}, AppID: row.AppID, InstallID: row.InstallID,
	}}
	for _, jobID := range row.RelatedJobIDs {
		target.RelatedJobs = append(target.RelatedJobs, permissions.Scope{Kind: permissions.KindRunnerJob, OrgID: row.OrgID, ID: jobID})
	}
	if err := target.ValidateShape(); err != nil {
		return permissions.NativeTarget{}, err
	}
	return target, nil
}
