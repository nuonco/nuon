package activities

import (
	"context"
)

type GetUnknownComponents struct {
}

// @temporal-gen activity
func (a *Activities) GetUnknownComponentIDs(ctx context.Context, req GetUnknownComponents) ([]string, error) {
	comps, err := a.getUnkownComponentIDs(ctx)
	if err != nil {
		return nil, err
	}

	return comps, nil
}

func (a *Activities) getUnkownComponentIDs(ctx context.Context) ([]string, error) {
	ids := make([]string, 0)
	res := a.db.WithContext(ctx).Raw("select id from public.components where type IS NULL ORDER BY created_at ASC;").Scan(&ids)
	if res.Error != nil {
		return nil, res.Error
	}

	return ids, nil
}
