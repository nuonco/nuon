package app

import (
	"database/sql/driver"
	"fmt"
)

// Now that we run third type of runner process, we need to distinguish which process is sending the heartbear
// NOTE(fd): we have to implement Scan/Value so the gorm ch plugin doesn't complain
type RunnerProcess string

func (rp RunnerProcess) Value() (driver.Value, error) {
	return string(rp), nil
}

func (rp *RunnerProcess) Scan(value any) error {
	if value == nil {
		*rp = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*rp = RunnerProcess(v)
	case []byte:
		*rp = RunnerProcess(v)
	default:
		return fmt.Errorf("cannot scan %T into RunnerProcess", value)
	}

	return nil
}

const (
	RunnerProcessMng     RunnerProcess = "mng"
	RunnerProcessInstall RunnerProcess = "install"
	RunnerProcessOrg     RunnerProcess = "org"
	RunnerProcessUknown  RunnerProcess = ""
)
