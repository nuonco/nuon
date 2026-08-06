// Package day2 defines the S3 contracts shared by the resident air-gap
// runner, the nuon-bundle CLI, and the portal: the dispatch mailbox
// (immutable requests, conditionally-created claims, terminal receipts), the
// per-run state layout, and the catalog of dispatchable refs.
//
// S3 is a mailbox, not a queue: dispatch is at-least-once and duplicate
// dispatches of the same dispatch ID resolve to the same receipt. All keys
// below are relative to the deployment's state prefix.
package day2

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/robfig/cron"
)

const SchemaVersion = 1

const (
	RequestsPrefix  = "dispatch/requests/"
	ClaimsPrefix    = "dispatch/claims/"
	ReceiptsPrefix  = "dispatch/receipts/"
	RunsPrefix      = "runs/"
	SchedulesPrefix = "schedules/"
	CatalogKey      = "day2/catalog.json"
)

func RequestKey(dispatchID string) string { return RequestsPrefix + dispatchID + ".json" }
func ClaimKey(dispatchID string) string   { return ClaimsPrefix + dispatchID + ".json" }
func ReceiptKey(dispatchID string) string { return ReceiptsPrefix + dispatchID + ".json" }
func RunStatusKey(runID string) string    { return RunsPrefix + runID + "/status.json" }
func RunStepLogKey(runID, stepID string) string {
	return RunsPrefix + runID + "/steps/" + stepID + "/logs.txt"
}
func RunStepResultKey(runID, stepID string) string {
	return RunsPrefix + runID + "/steps/" + stepID + "/result.json"
}
func ScheduleCursorKey(scheduleID string) string {
	return SchedulesPrefix + scheduleID + "/cursor.json"
}

var dispatchIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,128}$`)

// ValidateDispatchID keeps dispatch IDs safe to splice into S3 keys and file
// paths on both the writer and reader side.
func ValidateDispatchID(id string) error {
	if !dispatchIDPattern.MatchString(id) {
		return fmt.Errorf("dispatch ID %q must be 1-128 characters of [a-zA-Z0-9._-]", id)
	}
	return nil
}

const (
	SourcePortal = "portal"
	SourceCLI    = "cli"
	SourceCron   = "cron"
)

// Request is the immutable dispatch document a portal or CLI writes with
// If-None-Match:*. RequestedBy is display-only; CloudTrail S3 data events are
// the audit truth.
type Request struct {
	SchemaVersion int        `json:"schema_version"`
	DeploymentID  string     `json:"deployment_id"`
	BundleDigest  string     `json:"bundle_digest"`
	RefID         string     `json:"ref_id"`
	DispatchID    string     `json:"dispatch_id"`
	Source        string     `json:"source"`
	RequestedBy   string     `json:"requested_by,omitempty"`
	ScheduledAt   *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Claim is conditionally created (If-None-Match:*) by the runner before it
// executes a request; a claim older than its lease expiry may be retried by a
// successor runner with an If-Match takeover.
type Claim struct {
	DispatchID string    `json:"dispatch_id"`
	Owner      string    `json:"owner"`
	RunID      string    `json:"run_id"`
	Attempt    int       `json:"attempt"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

const (
	ReceiptStatusFinished = "finished"
	ReceiptStatusFailed   = "failed"
	// ReceiptStatusRejected marks a request the runner refused to execute:
	// unknown ref, bundle digest mismatch, or malformed document.
	ReceiptStatusRejected = "rejected"
)

// Receipt is terminal: it is written once, after the run's state is durable,
// and never updated.
type Receipt struct {
	DispatchID string    `json:"dispatch_id"`
	RefID      string    `json:"ref_id,omitempty"`
	RunID      string    `json:"run_id,omitempty"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason,omitempty"`
	FinishedAt time.Time `json:"finished_at"`
}

const (
	RefKindAction  = "action"
	RefKindDrift   = "drift"
	RefKindRunbook = "runbook"
)

// Catalog is published by the resident runner so portals and CLIs can list
// dispatchable refs without reading the bundle.
type Catalog struct {
	SchemaVersion int          `json:"schema_version"`
	DeploymentID  string       `json:"deployment_id"`
	BundleDigest  string       `json:"bundle_digest"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Refs          []CatalogRef `json:"refs"`
}

type CatalogRef struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Component    string `json:"component,omitempty"`
	CronSchedule string `json:"cron_schedule,omitempty"`
	Steps        int    `json:"steps,omitempty"`
}

const (
	RunStatusInProgress = "in-progress"
	RunStatusFinished   = "finished"
	RunStatusFailed     = "failed"
)

// RunStatus is the authoritative record of one day-2 run, stored at
// runs/<run-id>/status.json and mirrored to S3.
type RunStatus struct {
	RunID      string     `json:"run_id"`
	DispatchID string     `json:"dispatch_id"`
	RefID      string     `json:"ref_id"`
	RefKind    string     `json:"ref_kind"`
	RefName    string     `json:"ref_name"`
	Source     string     `json:"source"`
	Status     string     `json:"status"`
	Error      string     `json:"error,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Steps      []RunStep  `json:"steps"`
}

type RunStep struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	// JobID is the runtime job ID (run ID + template step ID) under which
	// this step executed; job logs are addressed by it.
	JobID      string       `json:"job_id,omitempty"`
	Status     string       `json:"status"`
	Error      string       `json:"error,omitempty"`
	StartedAt  *time.Time   `json:"started_at,omitempty"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
	Drift      *DriftResult `json:"drift,omitempty"`
}

// DriftResult classifies a drift run's fresh plan JSON. Drifted reports
// divergence from the bundle's frozen desired config.
type DriftResult struct {
	Drifted         bool   `json:"drifted"`
	ResourceChanges int    `json:"resource_changes"`
	OutputChanges   int    `json:"output_changes"`
	ResourceDrift   int    `json:"resource_drift"`
	Summary         string `json:"summary,omitempty"`
}

// ScheduleCursor records the last handled cron occurrence so a restarted
// runner neither backfills missed ticks nor re-fires the last one.
type ScheduleCursor struct {
	ScheduleID  string     `json:"schedule_id"`
	LastFiredAt *time.Time `json:"last_fired_at,omitempty"`
	// Skipped counts occurrences that were skipped because a run was still
	// active (overlap forbidden) or because the tick passed while the runner
	// was down (no backfill).
	Skipped       int        `json:"skipped,omitempty"`
	LastSkippedAt *time.Time `json:"last_skipped_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// OccurrenceID derives the deterministic dispatch ID for one cron occurrence,
// so every runner that observes the same tick claims the same dispatch.
func OccurrenceID(deploymentID, bundleDigest, scheduleID string, scheduledAt time.Time) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%d", deploymentID, bundleDigest, scheduleID, scheduledAt.UTC().Unix())
	return "occ-" + hex.EncodeToString(h.Sum(nil))[:20]
}

// EnvelopeDigest is the deployment's bundle-identity stand-in: the digest of
// the plan envelope bytes as loaded by the runner. Portals echo it back in
// dispatch requests so a request created against a different bundle is
// rejected instead of silently executing the wrong template.
func EnvelopeDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// NextAfter returns the next cron occurrence strictly after the given time,
// evaluated in UTC. The schedule must be a five-field cron expression.
func NextAfter(schedule string, after time.Time) (time.Time, error) {
	parsed, err := cron.ParseStandard(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron schedule %q: %w", schedule, err)
	}
	return parsed.Next(after.UTC()), nil
}
