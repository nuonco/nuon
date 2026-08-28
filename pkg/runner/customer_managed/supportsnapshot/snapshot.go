package supportsnapshot

import (
	"encoding/json"
	"time"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

const (
	SchemaVersion      = 1
	ManifestPath       = "manifest.json"
	RegistrationPath   = "installation-registration.json"
	SnapshotPath       = "snapshot.json"
	CollectionPath     = "collection-report.json"
	ArchiveContentType = "application/vnd.nuon.customer-managed-support-snapshot+zstd"
)

type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	CapturedAt    time.Time       `json:"captured_at"`
	Producer      Producer        `json:"producer"`
	Registration  string          `json:"registration_id"`
	BundleDigest  string          `json:"bundle_digest"`
	Entries       []ManifestEntry `json:"entries"`
}

type Producer struct {
	Name          string `json:"name"`
	Version       string `json:"version,omitempty"`
	RunnerVersion string `json:"runner_version,omitempty"`
}

type ManifestEntry struct {
	Path          string `json:"path"`
	MediaType     string `json:"media_type"`
	SchemaVersion int    `json:"schema_version"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
}

type CollectionReport struct {
	SchemaVersion int               `json:"schema_version"`
	Redaction     string            `json:"redaction_policy"`
	Included      []string          `json:"included"`
	Unavailable   map[string]string `json:"unavailable,omitempty"`
	Truncated     map[string]int64  `json:"truncated,omitempty"`
}

type Snapshot struct {
	SchemaVersion     int                                      `json:"schema_version"`
	CapturedAt        time.Time                                `json:"captured_at"`
	Registration      customermanaged.InstallationRegistration `json:"registration"`
	IncludeState      bool                                     `json:"include_state"`
	State             *CapturedState                           `json:"state,omitempty"`
	Runner            *customermanaged.RunnerHeartbeat         `json:"runner,omitempty"`
	Catalog           *operation.Catalog                       `json:"catalog,omitempty"`
	ActiveBundle      *operation.BundleInfo                    `json:"active_bundle,omitempty"`
	StagedBundle      *operation.BundleCandidate               `json:"staged_bundle,omitempty"`
	BundleHistory     []operation.BundleInfo                   `json:"bundle_history,omitempty"`
	Health            *customermanaged.HealthSnapshot          `json:"health,omitempty"`
	HealthTransitions []customermanaged.HealthTransition       `json:"health_transitions,omitempty"`
	CurrentInputs     *customermanaged.CapturedInputs          `json:"current_inputs,omitempty"`
	Roles             *customermanaged.CapturedRoles           `json:"roles,omitempty"`
	Runs              []Run                                    `json:"runs,omitempty"`
	Logs              []JobLog                                 `json:"logs,omitempty"`
	Collection        CollectionReport                         `json:"collection"`
}

type CapturedState struct {
	Status json.RawMessage `json:"status,omitempty" swaggertype:"object"`
	Report json.RawMessage `json:"report,omitempty" swaggertype:"object"`
}

type Run struct {
	RunID           string                     `json:"run_id"`
	DispatchID      string                     `json:"dispatch_id,omitempty"`
	RefID           string                     `json:"ref_id"`
	RefKind         string                     `json:"ref_kind"`
	RefName         string                     `json:"ref_name"`
	Source          string                     `json:"source"`
	Status          string                     `json:"status"`
	Error           string                     `json:"error,omitempty"`
	BundleDigest    string                     `json:"bundle_digest,omitempty"`
	PreviousRunID   string                     `json:"previous_run_id,omitempty"`
	StartedAt       time.Time                  `json:"started_at"`
	FinishedAt      *time.Time                 `json:"finished_at,omitempty"`
	Steps           []RunStep                  `json:"steps"`
	ResultDirective statestore.ResultDirective `json:"result_directive,omitempty"`
}

type RunStep struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Kind            string                     `json:"kind"`
	JobID           string                     `json:"job_id,omitempty"`
	Status          string                     `json:"status"`
	Error           string                     `json:"error,omitempty"`
	StartedAt       *time.Time                 `json:"started_at,omitempty"`
	FinishedAt      *time.Time                 `json:"finished_at,omitempty"`
	SourceRunID     string                     `json:"source_run_id,omitempty"`
	ResultDirective statestore.ResultDirective `json:"result_directive,omitempty"`
	Description     string                     `json:"status_description,omitempty"`
	Plan            *StepPlan                  `json:"plan,omitempty"`
	Drift           *operation.DriftResult     `json:"drift,omitempty"`
}

type StepPlan struct {
	Kind    string          `json:"kind"`
	Content json.RawMessage `json:"content" swaggertype:"object"`
}

type JobLog struct {
	JobID     string     `json:"job_id"`
	RunID     string     `json:"run_id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	Entries   []LogEntry `json:"entries"`
	Total     int        `json:"total"`
	Truncated bool       `json:"truncated,omitempty"`
}

type LogEntry struct {
	Time   *time.Time     `json:"time,omitempty"`
	Level  string         `json:"level,omitempty"`
	Msg    string         `json:"msg,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
}
