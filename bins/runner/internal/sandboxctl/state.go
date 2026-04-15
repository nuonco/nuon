package sandboxctl

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type ResponsePreset string

const (
	PresetDefault        ResponsePreset = "default"
	PresetSuccessSlow    ResponsePreset = "success_slow"
	PresetSuccessFast    ResponsePreset = "success_fast"
	PresetFailure        ResponsePreset = "failure"
	PresetFailureTimeout ResponsePreset = "failure_timeout"
	PresetFailurePanic   ResponsePreset = "failure_panic"
	PresetPartialFailure ResponsePreset = "partial_failure"
)

type JobTypeConfig struct {
	Preset       ResponsePreset         `json:"preset"`
	Duration     time.Duration          `json:"duration_ms"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Outputs      map[string]interface{} `json:"outputs,omitempty"`
	FaultRate    float64                `json:"fault_rate"`

	// API-driven fields (synced from centralized sandbox configs)
	LogLines        []string      `json:"log_lines,omitempty"`
	PlanContents    string        `json:"plan_contents,omitempty"`
	FailAtStep      string        `json:"fail_at_step,omitempty"`
	SleepDuration   time.Duration `json:"sleep_duration,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	TriggerShutdown bool          `json:"trigger_shutdown,omitempty"`
}

type State struct {
	mu sync.RWMutex

	JobTypes          map[string]*JobTypeConfig `json:"job_types"`
	DefaultDuration   time.Duration             `json:"default_duration"`
	PanicOnNextJob    bool                      `json:"panic_on_next_job"`
	ShutdownOnNextJob bool                      `json:"shutdown_on_next_job"`

	jobsExecuted  atomic.Int64
	jobsSucceeded atomic.Int64
	jobsFailed    atomic.Int64
}

type Stats struct {
	JobsExecuted  int64 `json:"jobs_executed"`
	JobsSucceeded int64 `json:"jobs_succeeded"`
	JobsFailed    int64 `json:"jobs_failed"`
}

type StateSnapshot struct {
	JobTypes          map[string]*JobTypeConfig `json:"job_types"`
	DefaultDuration   time.Duration             `json:"default_duration"`
	PanicOnNextJob    bool                      `json:"panic_on_next_job"`
	ShutdownOnNextJob bool                      `json:"shutdown_on_next_job"`
	Stats             Stats                     `json:"stats"`
}

func NewState(defaultDuration time.Duration) *State {
	s := &State{
		JobTypes:        make(map[string]*JobTypeConfig),
		DefaultDuration: defaultDuration,
	}
	s.initDefaults()
	return s
}

func (s *State) initDefaults() {
	for _, jt := range AllJobTypes() {
		s.JobTypes[jt] = &JobTypeConfig{
			Preset:   PresetDefault,
			Duration: s.DefaultDuration,
		}
	}
}

func (s *State) GetConfig(jobType string) JobTypeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cfg, ok := s.JobTypes[jobType]; ok {
		return *cfg
	}
	return JobTypeConfig{
		Preset:   PresetDefault,
		Duration: s.DefaultDuration,
	}
}

func (s *State) SetConfig(jobType string, cfg *JobTypeConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JobTypes[jobType] = cfg
}

func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.JobTypes = make(map[string]*JobTypeConfig)
	s.PanicOnNextJob = false
	s.ShutdownOnNextJob = false
	s.jobsExecuted.Store(0)
	s.jobsSucceeded.Store(0)
	s.jobsFailed.Store(0)
	s.initDefaults()
}

func (s *State) CheckAndClearPanic() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.PanicOnNextJob {
		s.PanicOnNextJob = false
		return true
	}
	return false
}

func (s *State) CheckAndClearShutdown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ShutdownOnNextJob {
		s.ShutdownOnNextJob = false
		return true
	}
	return false
}

func (s *State) SetPanic() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PanicOnNextJob = true
}

func (s *State) SetShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ShutdownOnNextJob = true
}

func (s *State) RecordResult(success bool) {
	s.jobsExecuted.Add(1)
	if success {
		s.jobsSucceeded.Add(1)
	} else {
		s.jobsFailed.Add(1)
	}
}

func (s *State) GetStats() Stats {
	return Stats{
		JobsExecuted:  s.jobsExecuted.Load(),
		JobsSucceeded: s.jobsSucceeded.Load(),
		JobsFailed:    s.jobsFailed.Load(),
	}
}

func (s *State) Snapshot() StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jt := make(map[string]*JobTypeConfig, len(s.JobTypes))
	for k, v := range s.JobTypes {
		cpy := *v
		jt[k] = &cpy
	}

	return StateSnapshot{
		JobTypes:          jt,
		DefaultDuration:   s.DefaultDuration,
		PanicOnNextJob:    s.PanicOnNextJob,
		ShutdownOnNextJob: s.ShutdownOnNextJob,
		Stats:             s.GetStats(),
	}
}

func (s *State) GetAllConfigs() map[string]*JobTypeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jt := make(map[string]*JobTypeConfig, len(s.JobTypes))
	for k, v := range s.JobTypes {
		cpy := *v
		jt[k] = &cpy
	}
	return jt
}

// SyncFromAPI overwrites local state with configs fetched from the centralized API.
func (s *State) SyncFromAPI(configs []*nuonrunner.SandboxConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cfg := range configs {
		var logLines []string
		if len(cfg.LogLines) > 0 {
			_ = json.Unmarshal(cfg.LogLines, &logLines)
		}

		var outputs map[string]interface{}
		if len(cfg.Outputs) > 0 {
			_ = json.Unmarshal(cfg.Outputs, &outputs)
		}

		preset := ResponsePreset(cfg.Preset)
		if preset == "" {
			preset = PresetDefault
		}

		duration := cfg.Duration
		if duration == 0 {
			duration = s.DefaultDuration
		}

		s.JobTypes[cfg.JobType] = &JobTypeConfig{
			Preset:          preset,
			Duration:        duration,
			FaultRate:       cfg.FaultRate,
			ErrorMessage:    cfg.ErrorMessage,
			LogLines:        logLines,
			PlanContents:    cfg.PlanContents,
			FailAtStep:      cfg.FailAtStep,
			SleepDuration:   cfg.SleepDuration,
			Timeout:         cfg.Timeout,
			TriggerShutdown: cfg.TriggerShutdown,
			Outputs:         outputs,
		}
	}
}
