package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type ResourceSet struct {
	kind  permissions.ResourceKind
	query func(*gorm.DB, string) *gorm.DB
}

func DefineResources(kind permissions.ResourceKind, query func(*gorm.DB, string) *gorm.DB) ResourceSet {
	return ResourceSet{kind: kind, query: query}
}

func (r ResourceSet) Query(db *gorm.DB, orgID string) *gorm.DB {
	return r.query(db, orgID)
}

func (r ResourceSet) In(db *gorm.DB, orgID string) (*Resources, error) {
	return NewResources(r.query(db, orgID), r.kind, orgID)
}

type Selection struct {
	ids *gorm.DB
}

func (s *Selection) IDs() *gorm.DB {
	return s.ids.Session(&gorm.Session{})
}

func (s *Selection) Scope(columns ...clause.Column) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		column := clause.Column{Table: clause.CurrentTable, Name: "id"}
		if len(columns) > 1 {
			db.AddError(fmt.Errorf("authorization selection requires one ID column"))
			return db
		}
		if len(columns) == 1 {
			column = columns[0]
		}
		return db.Where(clause.Expr{SQL: "? IN (?)", Vars: []any{column, s.IDs()}})
	}
}

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
