package jobloop

import (
	"time"
)

// Healthcheck
//
// On every healthcheck run, we update the relevant values in the Healthcheck struct.
// These are returned as outputs by the healthcheck job handler. We can then use them
// on the ctl-api side to determine the runner health and take action.

type Healthcheck struct {
	StartTime           time.Time
	StopTime            time.Time
	LatestHealthcheckAt time.Time
}

func (j *jobLoop) GetHealthcheck() (Healthcheck, string) {
	return j.healthcheck, string(j.jobGroup)
}

func (j *jobLoop) setStarted() error {
	j.healthcheck.StartTime = time.Now()
	return nil
}

func (j *jobLoop) setStopped() error {
	j.healthcheck.StopTime = time.Now()
	return nil
}

func (j *jobLoop) SetLatestHealthcheckAt() error {
	j.healthcheck.LatestHealthcheckAt = time.Now()
	return nil
}

func (j *jobLoop) TimeSinceLastHealthcheck() time.Duration {
	if j.healthcheck.LatestHealthcheckAt.IsZero() {
		return time.Duration(0)
	}
	return time.Now().Sub(j.healthcheck.LatestHealthcheckAt)
}

// func (j *jobLoop) setLatestJobRun(start time.Time) error {
// 	j.healthcheck.LatestJobRunAt = start
// 	return nil
// }

// func (j *jobLoop) setLatestJobDuration(d time.Duration) error {
// 	j.healthcheck.LatestJobRunDuration = d
// 	return nil
// }

// func (j *jobLoop) timeSinceLastJobRun() time.Duration {
// 	if j.healthcheck.LatestJobRunAt.IsZero() {
// 		return time.Now().Sub(j.healthcheck.LatestJobRunAt)
// 	}
// 	return time.Duration(0)
// }
