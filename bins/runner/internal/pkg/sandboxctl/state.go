package sandboxctl

import (
	"sync"
	"time"
)

// JobCategory represents the type of sandbox job.
type JobCategory string

const (
	CategoryTerraform  JobCategory = "terraform"
	CategoryHelm       JobCategory = "helm"
	CategoryKubernetes JobCategory = "kubernetes"
)

// AllCategories returns all supported job categories.
func AllCategories() []JobCategory {
	return []JobCategory{CategoryTerraform, CategoryHelm, CategoryKubernetes}
}

// ResponseVariant selects which fixture to use for a given category.
type ResponseVariant string

const (
	VariantDefault ResponseVariant = "default"
	VariantEmpty   ResponseVariant = "empty"
	VariantLarge   ResponseVariant = "large"
)

// FailureMode controls how sandbox jobs fail.
type FailureMode string

const (
	FailureModeNone     FailureMode = "none"
	FailureModeError    FailureMode = "error"
	FailureModePanic    FailureMode = "panic"
	FailureModeShutdown FailureMode = "shutdown"
)

// StateSnapshot is the JSON-serializable state of the sandbox controller.
type StateSnapshot struct {
	Variants      map[JobCategory]ResponseVariant `json:"variants"`
	FailureMode   FailureMode                     `json:"failure_mode"`
	JobDuration   string                          `json:"job_duration"`
	FaultsEnabled bool                            `json:"faults_enabled"`
}

// State holds the current sandbox control configuration.
type State struct {
	mu            sync.RWMutex
	variants      map[JobCategory]ResponseVariant
	failureMode   FailureMode
	jobDuration   time.Duration
	faultsEnabled bool
}

// NewState creates a State with default values.
func NewState() *State {
	return &State{
		variants: map[JobCategory]ResponseVariant{
			CategoryTerraform:  VariantDefault,
			CategoryHelm:       VariantDefault,
			CategoryKubernetes: VariantDefault,
		},
		failureMode:   FailureModeNone,
		jobDuration:   0,
		faultsEnabled: false,
	}
}

// Snapshot returns a point-in-time copy of the state.
func (s *State) Snapshot() StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	variants := make(map[JobCategory]ResponseVariant, len(s.variants))
	for k, v := range s.variants {
		variants[k] = v
	}

	dur := ""
	if s.jobDuration > 0 {
		dur = s.jobDuration.String()
	}

	return StateSnapshot{
		Variants:      variants,
		FailureMode:   s.failureMode,
		JobDuration:   dur,
		FaultsEnabled: s.faultsEnabled,
	}
}

// Apply updates the full state from a snapshot.
func (s *State) Apply(snap StateSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for cat, v := range snap.Variants {
		s.variants[cat] = v
	}
	s.failureMode = snap.FailureMode
	if snap.JobDuration != "" {
		if d, err := time.ParseDuration(snap.JobDuration); err == nil {
			s.jobDuration = d
		}
	} else {
		s.jobDuration = 0
	}
	s.faultsEnabled = snap.FaultsEnabled
}

// SetVariant sets the response variant for a specific category.
func (s *State) SetVariant(cat JobCategory, v ResponseVariant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variants[cat] = v
}

// SetFailureMode sets the failure mode.
func (s *State) SetFailureMode(fm FailureMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failureMode = fm
}

// GetVariant returns the variant for a category.
func (s *State) GetVariant(cat JobCategory) ResponseVariant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.variants[cat]; ok {
		return v
	}
	return VariantDefault
}

// GetFailureMode returns the current failure mode.
func (s *State) GetFailureMode() FailureMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.failureMode
}

// GetJobDuration returns the job duration override (0 means use config default).
func (s *State) GetJobDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobDuration
}

// GetFaultsEnabled returns whether random fault injection is enabled.
func (s *State) GetFaultsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.faultsEnabled
}
