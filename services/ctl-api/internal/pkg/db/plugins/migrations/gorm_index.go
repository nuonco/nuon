package migrations

import "gorm.io/gorm"

func toIndexMap[T gorm.Index](indexes []T) map[string]gorm.Index {
	vals := make(map[string]gorm.Index, 0)
	for _, val := range indexes {
		idx := val
		vals[idx.Name()] = idx
	}

	return vals
}

type gormIndex struct {
	idx   Index
	table string
}

func (g gormIndex) Name() string {
	return g.idx.Name
}

func (g gormIndex) Columns() []string {
	return g.idx.Columns
}

func (g gormIndex) Option() string {
	return g.idx.Option
}

func (g gormIndex) PrimaryKey() (bool, bool) {
	return g.idx.PrimaryValue.Bool, g.idx.PrimaryValue.Valid
}

func (g gormIndex) Table() string {
	return g.table
}

func (g gormIndex) Unique() (bool, bool) {
	return g.idx.UniqueValue.Bool, g.idx.UniqueValue.Valid
}

var _ gorm.Index = (*gormIndex)(nil)
