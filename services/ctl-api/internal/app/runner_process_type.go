package app

import (
	"database/sql/driver"
	"fmt"
)

// RunnerProcessType distinguishes which process type is sending heartbeats/health checks.
// NOTE(fd): we have to implement Scan/Value so the gorm ch plugin doesn't complain
type RunnerProcessType string

func (rp RunnerProcessType) Value() (driver.Value, error) {
	return string(rp), nil
}

func (rp *RunnerProcessType) Scan(value any) error {
	if value == nil {
		*rp = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*rp = RunnerProcessType(v)
	case []byte:
		*rp = RunnerProcessType(v)
	default:
		return fmt.Errorf("cannot scan %T into RunnerProcessType", value)
	}

	return nil
}

const (
	RunnerProcessTypeMng     RunnerProcessType = "mng"
	RunnerProcessTypeInstall RunnerProcessType = "install"
	RunnerProcessTypeOrg     RunnerProcessType = "org"
	RunnerProcessTypeUnknown RunnerProcessType = ""
)
