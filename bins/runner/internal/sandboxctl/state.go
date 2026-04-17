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

	// API-driven fields (synced from centralized sandbox configs)
	LogLines            []string      `json:"log_lines,omitempty"`
	PlanContents        string        `json:"plan_contents,omitempty"`
	PlanDisplayContents string        `json:"plan_display_contents,omitempty"`
	StateJSON           string        `json:"state_json,omitempty"`
	FailAtStep          string        `json:"fail_at_step,omitempty"`
	SleepDuration       time.Duration `json:"sleep_duration,omitempty"`
	Timeout             time.Duration `json:"timeout,omitempty"`
	TriggerShutdown     bool          `json:"trigger_shutdown,omitempty"`
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

// configKey returns the composite key for a job type and operation.
func configKey(jobType, operation string) string {
	if operation == "" {
		return jobType
	}
	return jobType + "|" + operation
}

func (s *State) GetConfig(jobType string) JobTypeConfig {
	return s.GetConfigForOperation(jobType, "")
}

// GetConfigForOperation returns the config for a job type and operation.
// It first checks for an exact match on job_type+operation, then falls back
// to the job_type-only config (operation="").
func (s *State) GetConfigForOperation(jobType, operation string) JobTypeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try exact match with operation first
	if operation != "" {
		if cfg, ok := s.JobTypes[configKey(jobType, operation)]; ok {
			return *cfg
		}
	}

	// Fall back to job-type-only config
	if cfg, ok := s.JobTypes[configKey(jobType, "")]; ok {
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
		if !cfg.Enabled {
			continue
		}

		s.JobTypes[configKey(cfg.JobType, cfg.Operation)] = configFromAPI(cfg, s.DefaultDuration)
	}
}

// SyncSingleFromAPI stores a single config fetched from the API for a specific job.
func (s *State) SyncSingleFromAPI(cfg *nuonrunner.SandboxConfig) {
	if cfg == nil || !cfg.Enabled {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.JobTypes[configKey(cfg.JobType, cfg.Operation)] = configFromAPI(cfg, s.DefaultDuration)
}

func configFromAPI(cfg *nuonrunner.SandboxConfig, defaultDuration time.Duration) *JobTypeConfig {
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
		duration = defaultDuration
	}

	return &JobTypeConfig{
		Preset:              preset,
		Duration:            duration,
		ErrorMessage:        cfg.ErrorMessage,
		LogLines:            logLines,
		PlanContents:        cfg.PlanContents,
		PlanDisplayContents: cfg.PlanDisplayContents,
		StateJSON:           cfg.StateJSON,
		FailAtStep:          cfg.FailAtStep,
		SleepDuration:       cfg.SleepDuration,
		Timeout:             cfg.Timeout,
		TriggerShutdown:     cfg.TriggerShutdown,
		Outputs:             outputs,
	}
}
