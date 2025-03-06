package generics

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func BatchProcessing[T any](ctx context.Context, batchSize int, query *gorm.DB, processBatch func([]*T) error) error {
	var items []*T
	offset := 0
	for {
		result := query.
			Offset(offset).
			Limit(batchSize).
			Find(&items)
		if result.Error != nil {
			return fmt.Errorf("unable to fetch items: %w", result.Error)
		}
		if len(items) == 0 {
			break
		}
		if err := processBatch(items); err != nil {
			return err
		}
		offset += batchSize
	}
	return nil
}
