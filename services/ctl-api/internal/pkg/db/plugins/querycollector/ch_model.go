package querycollector

import "time"

// CHQueryRecord is the GORM model for the ClickHouse `queries` table.
type CHQueryRecord struct {
	Table        string    `gorm:"column:table"`
	Operation    string    `gorm:"column:operation"`
	SQL          string    `gorm:"column:sql"`
	DurationMS   float64   `gorm:"column:duration_ms"`
	RowsAffected int64     `gorm:"column:rows_affected"`
	ResponseSize int       `gorm:"column:response_size"`
	PreloadCount int       `gorm:"column:preload_count"`
	Timestamp    time.Time `gorm:"column:timestamp"`
	Error        string    `gorm:"column:error"`
	Caller       string    `gorm:"column:caller"`
	CallerURL    string    `gorm:"column:caller_url"`
	DBType       string    `gorm:"column:db_type"`
	Source       string    `gorm:"column:source"`
	Endpoint     string    `gorm:"column:endpoint"`
	ProcessID    string    `gorm:"column:process_id"`
}

func (CHQueryRecord) TableName() string { return "queries" }
