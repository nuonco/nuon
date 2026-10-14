package orgscope

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// associationScopeKey marks the statement context once the parent query has been
// considered. Preload and association saves copy that context, so they skip scoping.
type associationScopeKey struct{}

// skipKey marks a context the auth listener owns. Those queries are not org-scoped.
type skipKey struct{}

// Skip returns a context the org-scope plugin ignores.
func Skip(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipKey{}, true)
}

func New() gorm.Plugin {
	return &plugin{}
}

type plugin struct{}

func (p *plugin) Name() string {
	return "org-scope"
}

func (p *plugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Query().Before("gorm:query").Register("org_scope:query", p.scope); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:save_before_associations").Register("org_scope:update", p.scope); err != nil {
		return err
	}
	return db.Callback().Delete().Before("gorm:delete_before_associations").Register("org_scope:delete", p.scope)
}

func (p *plugin) scope(tx *gorm.DB) {
	if tx.Statement == nil || tx.Statement.Context == nil {
		return
	}
	if tx.Statement.Context.Value(skipKey{}) != nil || tx.Statement.Context.Value(associationScopeKey{}) != nil {
		return
	}
	orgID := keys.OrgIDFromContext(tx.Statement.Context)
	if orgID == "" {
		return
	}
	tx.Statement.Context = context.WithValue(tx.Statement.Context, associationScopeKey{}, true)
	if !hasOrgForeignKey(tx.Statement.Schema) || whereScopedToOrg(tx.Statement, orgID) {
		return
	}

	tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{
			Column: clause.Column{Table: clause.CurrentTable, Name: "org_id"},
			Value:  orgID,
		},
	}})
}

func hasOrgForeignKey(s *schema.Schema) bool {
	if s == nil {
		return false
	}
	for _, rel := range s.Relationships.BelongsTo {
		if rel.FieldSchema == nil || rel.FieldSchema.Name != "Org" {
			continue
		}
		for _, ref := range rel.References {
			if ref.ForeignKey != nil && ref.ForeignKey.DBName == "org_id" {
				return true
			}
		}
	}
	return false
}
