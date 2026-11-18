package authz

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (r *Resources) AuthorizedIDs(ctx context.Context, catalog *permissions.Catalog, input permissions.Query) (*gorm.DB, error) {
	if input.Kind != r.kind || input.OrgID != r.orgID {
		return nil, fmt.Errorf("permission query must match the resource kind and org")
	}
	grants, ceiling, err := catalog.QueryScopes(input)
	if err != nil {
		return nil, err
	}
	query := r.query.WithContext(ctx).Select("permission_target.id").Where(scopePredicate(input.Kind, grants))
	if ceiling != nil {
		query = query.Where(scopePredicate(input.Kind, ceiling))
	}
	return query, nil
}

func scopePredicate(kind permissions.ResourceKind, scopes []permissions.GrantScope) clause.Expression {
	alternatives := make([]clause.Expression, 0, len(scopes))
	for _, scope := range scopes {
		column := ""
		switch scope.Kind {
		case permissions.KindOrg:
			column = "org_id"
		case permissions.KindApp:
			column = "app_id"
		case permissions.KindInstall:
			column = "install_id"
		case permissions.KindRunnerJob:
			if kind == permissions.KindLogStream {
				if slices.Contains(scope.IDs, "*") {
					alternatives = append(alternatives, clause.Expr{SQL: "jsonb_array_length(permission_target.related_job_ids) > 0"})
				} else {
					alternatives = append(alternatives, clause.Expr{
						SQL: "EXISTS (SELECT 1 FROM jsonb_array_elements_text(permission_target.related_job_ids) AS related_job(id) WHERE related_job.id IN ?)", Vars: []any{scope.IDs},
					})
				}
				continue
			}
			if kind == permissions.KindRunnerJob {
				column = "id"
			}
		default:
			if kind == scope.Kind {
				column = "id"
			}
		}
		if column == "" {
			continue
		}
		field := clause.Column{Table: "permission_target", Name: column}
		if slices.Contains(scope.IDs, "*") {
			alternatives = append(alternatives, clause.Neq{Column: field, Value: ""})
		} else {
			ids := make([]any, len(scope.IDs))
			for i, id := range scope.IDs {
				ids[i] = id
			}
			alternatives = append(alternatives, clause.IN{Column: field, Values: ids})
		}
	}
	if len(alternatives) == 0 {
		return clause.Expr{SQL: "FALSE"}
	}
	return clause.Or(alternatives...)
}
