package jobloop

import (
	"time"
)

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
	return time.Since(j.healthcheck.LatestHealthcheckAt)
}
