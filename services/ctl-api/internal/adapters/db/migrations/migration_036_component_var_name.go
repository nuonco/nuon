package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration036ComponentVarNames(ctx context.Context) error {
	// handle all deleted records so we can set a constraint
	sql := `
  UPDATE components SET var_name='' WHERE deleted_at > 0;
  `
	if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	var comps []*app.Component
	res := a.db.WithContext(ctx).
		Find(&comps)
	if res.Error != nil {
		return res.Error
	}

	for _, comp := range comps {
		res := a.db.WithContext(ctx).
			Model(&comp).
			Updates(app.Component{
				VarName: comp.Name,
			})
		if res.Error != nil {
			return fmt.Errorf("error updating components to set var name")
		}

	}

	return nil
}
