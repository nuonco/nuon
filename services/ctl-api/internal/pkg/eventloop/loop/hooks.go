package loop

import (
	"context"

	"gorm.io/gorm"
)

func CheckExists[Model any](ctx context.Context, db *gorm.DB, id string) (bool, error) {
	var count int64
	var typ Model

	db.WithContext(ctx).
		Model(typ).
		Where("id = ?", id).
		Count(&count)

	return count > 0, nil
}
