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
	RunnerProcessTypeBuild   RunnerProcessType = "build"
	RunnerProcessTypeOrg     RunnerProcessType = "org"
	RunnerProcessTypeUnknown RunnerProcessType = ""
)

// HeartBeatProcessForRunnerGroupType maps a RunnerGroupType to the RunnerProcessType
// used for heartbeat lookups. Install runner groups use the "mng" process type,
// org runner groups use the "org" process type.
func HeartBeatProcessForRunnerGroupType(gt RunnerGroupType) RunnerProcessType {
	switch gt {
	case RunnerGroupTypeInstall:
		return RunnerProcessTypeMng
	case RunnerGroupTypeOrg:
		return RunnerProcessTypeOrg
	default:
		return RunnerProcessTypeUnknown
	}
}
