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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron"
)

const SchemaVersion = 1

const (
	RequestsPrefix        = "dispatch/requests/"
	ClaimsPrefix          = "dispatch/claims/"
	ReceiptsPrefix        = "dispatch/receipts/"
	RunsPrefix            = "runs/"
	SchedulesPrefix       = "schedules/"
	CatalogKey            = "day2/catalog.json"
	BundleKey             = "day2/bundle.json"
	CandidateKey          = "day2/candidate.json"
	StagedCandidateKey    = "day2/staged-candidate.json"
	StackCandidateKey     = "day2/stack-candidate.json"
	CandidateStagesPrefix = "day2/candidates/"
	BundlesPrefix         = "day2/bundles/"
	JobPlansPrefix        = "job-plans/"
)

func RequestKey(dispatchID string) string { return RequestsPrefix + dispatchID + ".json" }
func ClaimKey(dispatchID string) string   { return ClaimsPrefix + dispatchID + ".json" }
func ReceiptKey(dispatchID string) string { return ReceiptsPrefix + dispatchID + ".json" }
func RunStatusKey(runID string) string    { return RunsPrefix + runID + "/status.json" }
func JobPlanKey(jobID string) string      { return JobPlansPrefix + jobID + ".json" }
func RunStepLogKey(runID, stepID string) string {
	return RunsPrefix + runID + "/steps/" + stepID + "/logs.txt"
}
func RunStepResultKey(runID, stepID string) string {
	return RunsPrefix + runID + "/steps/" + stepID + "/result.json"
}
func ScheduleCursorKey(scheduleID string) string {
	return SchedulesPrefix + scheduleID + "/cursor.json"
}

// BundleHistoryKey addresses the append-once activation record for one bundle
// digest. Digests contain a ":" separator, which is unsafe in file paths, so
// it is flattened to "-".
func BundleHistoryKey(digest string) string {
	return BundlesPrefix + strings.ReplaceAll(digest, ":", "-") + ".json"
}

func CandidateApprovalKey(digest string) string {
	return "day2/candidates/" + strings.ReplaceAll(digest, ":", "-") + "/approval.json"
}

func CandidateStageKey(digest string, stagedAt time.Time) string {
	return CandidateStagesPrefix + strings.ReplaceAll(digest, ":", "-") + "/staged/" + strconv.FormatInt(stagedAt.UTC().UnixNano(), 10) + ".json"
}

func CandidateDismissalKey(digest string, dismissedAt time.Time) string {
	return CandidateStagesPrefix + strings.ReplaceAll(digest, ":", "-") + "/dismissed/" + strconv.FormatInt(dismissedAt.UTC().UnixNano(), 10) + ".json"
}

func CandidateSandboxApprovalKey(digest string) string {
	return "day2/candidates/" + strings.ReplaceAll(digest, ":", "-") + "/sandbox-approval.json"
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
	SchemaVersion       int        `json:"schema_version"`
	DeploymentID        string     `json:"deployment_id"`
	BundleDigest        string     `json:"bundle_digest"`
	RefID               string     `json:"ref_id"`
	RefKind             string     `json:"ref_kind,omitempty"`
	RunID               string     `json:"run_id,omitempty"`
	CandidateArchiveKey string     `json:"candidate_archive_key,omitempty"`
	CandidateRecordKey  string     `json:"candidate_record_key,omitempty"`
	PlanStepIDs         []string   `json:"plan_step_ids,omitempty"`
	DispatchID          string     `json:"dispatch_id"`
	Source              string     `json:"source"`
	RequestedBy         string     `json:"requested_by,omitempty"`
	ScheduledAt         *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
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
	RefKindAction     = "action"
	RefKindDrift      = "drift"
	RefKindRunbook    = "runbook"
	RefKindBundlePlan = "bundle-plan"
)

func (r Request) ValidateBundlePlan() error {
	if r.RefKind != RefKindBundlePlan {
		return fmt.Errorf("ref kind must be %q", RefKindBundlePlan)
	}
	if err := ValidateDispatchID(r.RunID); err != nil {
		return fmt.Errorf("invalid run ID: %w", err)
	}
	if r.BundleDigest == "" || r.CandidateArchiveKey == "" || r.CandidateRecordKey == "" {
		return fmt.Errorf("bundle digest, candidate archive key, and candidate record key are required")
	}
	if len(r.PlanStepIDs) == 0 {
		return fmt.Errorf("at least one candidate plan step ID is required")
	}
	seen := make(map[string]bool, len(r.PlanStepIDs))
	for _, id := range r.PlanStepIDs {
		if err := ValidateDispatchID(id); err != nil {
			return fmt.Errorf("invalid plan step ID: %w", err)
		}
		if seen[id] {
			return fmt.Errorf("duplicate plan step ID %q", id)
		}
		seen[id] = true
	}
	return nil
}

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
	StepStatusDiscarded = "discarded"
)

// RunStatus is the authoritative record of one day-2 run, stored at
// runs/<run-id>/status.json and mirrored to S3.
type RunStatus struct {
	RunID        string     `json:"run_id"`
	DispatchID   string     `json:"dispatch_id"`
	RefID        string     `json:"ref_id"`
	RefKind      string     `json:"ref_kind"`
	RefName      string     `json:"ref_name"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	Error        string     `json:"error,omitempty"`
	BundleDigest string     `json:"bundle_digest,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	Steps        []RunStep  `json:"steps"`
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
	// Resources lists every resource in the plan with its change verdict,
	// changed resources first, capped at MaxDriftResources so run status
	// files stay small; the raw plan at JobPlanKey holds the full detail.
	Resources          []DriftResourceChange `json:"resources,omitempty"`
	ResourcesTruncated bool                  `json:"resources_truncated,omitempty"`
}

// MaxDriftResources bounds the per-resource list embedded in a run status.
const MaxDriftResources = 500

type DriftResourceChange struct {
	Address string `json:"address"`
	// Action is create, update, destroy, replace, or noop.
	Action string `json:"action"`
	// Drifted marks changes correlated with out-of-band resource drift.
	Drifted bool `json:"drifted,omitempty"`
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

const (
	BundleContentKindComponent    = "component"
	BundleContentKindSandbox      = "sandbox"
	BundleContentKindImage        = "image"
	BundleContentKindAction       = "action"
	BundleContentKindRunbook      = "runbook"
	BundleContentKindStackAsset   = "stack-asset"
	BundleContentKindRunnerBinary = "runner-binary"
	BundleContentKindRunnerImage  = "runner-image"
)

// BundleInfo is published by the resident runner when it activates a bundle:
// the active pointer lives at BundleKey and is rewritten on every activation,
// while an append-once copy at BundleHistoryKey(digest) preserves each
// digest's first activation for portal history.
type BundleInfo struct {
	SchemaVersion int    `json:"schema_version"`
	DeploymentID  string `json:"deployment_id"`
	BundleDigest  string `json:"bundle_digest"`
	// ArchiveDigest is the sha256 of the .tar.zst transport archive the
	// runner extracted, when the runner performed the extraction itself;
	// pre-extracted bundle directories leave it empty.
	ArchiveDigest string             `json:"archive_digest,omitempty"`
	ActivatedAt   time.Time          `json:"activated_at"`
	Target        *BundleTarget      `json:"target,omitempty"`
	Verification  BundleVerification `json:"verification"`
	TotalSize     int64              `json:"total_size,omitempty"`
	Contents      []BundleContent    `json:"contents,omitempty"`
}

const (
	BundleChangeAdded     = "added"
	BundleChangeChanged   = "changed"
	BundleChangeRemoved   = "removed"
	BundleChangeUnchanged = "unchanged"
)

type BundleCandidate struct {
	SchemaVersion  int                     `json:"schema_version"`
	PreviousDigest string                  `json:"previous_digest"`
	StagedAt       time.Time               `json:"staged_at"`
	ArchiveName    string                  `json:"archive_name,omitempty"`
	ArchiveSize    int64                   `json:"archive_size,omitempty"`
	Bundle         BundleInfo              `json:"bundle"`
	Changes        []BundleChange          `json:"changes"`
	Deployment     *BundleDeploymentAssets `json:"deployment,omitempty"`
}

type BundleCandidateDismissal struct {
	SchemaVersion int       `json:"schema_version"`
	BundleDigest  string    `json:"bundle_digest"`
	DismissedAt   time.Time `json:"dismissed_at"`
	RequestedBy   string    `json:"requested_by"`
}

type BundleDeploymentAssets struct {
	StackTemplateURL   string `json:"stack_template_url"`
	CandidateBundleKey string `json:"candidate_bundle_key"`
	TargetBundleKey    string `json:"target_bundle_key"`
}

type StackCandidate struct {
	SchemaVersion      int           `json:"schema_version"`
	BundleDigest       string        `json:"bundle_digest"`
	StackName          string        `json:"stack_name"`
	ChangeSetName      string        `json:"change_set_name"`
	ChangeSetARN       string        `json:"change_set_arn,omitempty"`
	TemplateURL        string        `json:"template_url"`
	CandidateBundleKey string        `json:"candidate_bundle_key"`
	TargetBundleKey    string        `json:"target_bundle_key"`
	CandidateRecordKey string        `json:"candidate_record_key,omitempty"`
	Status             string        `json:"status"`
	ExecutionStatus    string        `json:"execution_status"`
	StatusReason       string        `json:"status_reason,omitempty"`
	NoOp               bool          `json:"no_op,omitempty"`
	Changes            []StackChange `json:"changes,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	StackAppliedAt     *time.Time    `json:"stack_applied_at,omitempty"`
	RunnerActivatedAt  *time.Time    `json:"runner_activated_at,omitempty"`
	InstanceRefreshID  string        `json:"instance_refresh_id,omitempty"`
}

type StackChange struct {
	Action                   string                `json:"action"`
	LogicalResourceID        string                `json:"logical_resource_id"`
	ResourceType             string                `json:"resource_type"`
	Replacement              string                `json:"replacement,omitempty"`
	Scope                    []string              `json:"scope,omitempty"`
	Details                  []StackChangeDetail   `json:"details,omitempty"`
	PropertyChanges          []StackPropertyChange `json:"property_changes,omitempty"`
	PropertyChangesCaptured  bool                  `json:"property_changes_captured,omitempty"`
	PropertyChangesTruncated bool                  `json:"property_changes_truncated,omitempty"`
}

type StackChangeDetail struct {
	Attribute          string `json:"attribute,omitempty"`
	Name               string `json:"name,omitempty"`
	RequiresRecreation string `json:"requires_recreation,omitempty"`
	Evaluation         string `json:"evaluation,omitempty"`
	ChangeSource       string `json:"change_source,omitempty"`
	CausingEntity      string `json:"causing_entity,omitempty"`
}

type StackPropertyChange struct {
	Path   string `json:"path"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type BundleChange struct {
	Kind                         string                   `json:"kind"`
	Name                         string                   `json:"name"`
	Detail                       string                   `json:"detail,omitempty"`
	Change                       string                   `json:"change"`
	PreviousDigest               string                   `json:"previous_digest,omitempty"`
	CandidateDigest              string                   `json:"candidate_digest,omitempty"`
	PreviousConfig               string                   `json:"previous_config_digest,omitempty"`
	CandidateConfig              string                   `json:"candidate_config_digest,omitempty"`
	PreviousComponentDefinition  map[string]any           `json:"previous_component_definition,omitempty"`
	CandidateComponentDefinition map[string]any           `json:"candidate_component_definition,omitempty"`
	PreviousActionDefinition     *BundleActionDefinition  `json:"previous_action_definition,omitempty"`
	CandidateActionDefinition    *BundleActionDefinition  `json:"candidate_action_definition,omitempty"`
	PreviousRunbookDefinition    *BundleRunbookDefinition `json:"previous_runbook_definition,omitempty"`
	CandidateRunbookDefinition   *BundleRunbookDefinition `json:"candidate_runbook_definition,omitempty"`
	PlanStepID                   string                   `json:"plan_step_id,omitempty"`
	ApplyStepID                  string                   `json:"apply_step_id,omitempty"`
}

func CompareBundleContents(previous, candidate BundleInfo) []BundleChange {
	type key struct{ kind, name string }
	before := make(map[key]BundleContent, len(previous.Contents))
	after := make(map[key]BundleContent, len(candidate.Contents))
	for _, content := range previous.Contents {
		before[key{content.Kind, content.Name}] = content
	}
	for _, content := range candidate.Contents {
		after[key{content.Kind, content.Name}] = content
	}
	keys := make([]key, 0, len(before)+len(after))
	seen := map[key]bool{}
	for item := range before {
		keys = append(keys, item)
		seen[item] = true
	}
	for item := range after {
		if !seen[item] {
			keys = append(keys, item)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].kind == keys[j].kind {
			return keys[i].name < keys[j].name
		}
		return keys[i].kind < keys[j].kind
	})
	changes := make([]BundleChange, 0, len(keys))
	for _, item := range keys {
		old, hadOld := before[item]
		current, hasCurrent := after[item]
		change := BundleChange{
			Kind: item.kind, Name: item.name, Detail: current.Detail,
			PreviousDigest: old.Digest, CandidateDigest: current.Digest,
			PreviousConfig: old.ConfigDigest, CandidateConfig: current.ConfigDigest,
			PreviousComponentDefinition: old.ComponentDefinition, CandidateComponentDefinition: current.ComponentDefinition,
			PreviousActionDefinition: old.ActionDefinition, CandidateActionDefinition: current.ActionDefinition,
			PreviousRunbookDefinition: old.RunbookDefinition, CandidateRunbookDefinition: current.RunbookDefinition,
		}
		switch {
		case !hadOld:
			change.Change = BundleChangeAdded
		case !hasCurrent:
			change.Change = BundleChangeRemoved
			change.Detail = old.Detail
		case old.Digest != current.Digest || old.ConfigDigest != current.ConfigDigest || old.Detail != current.Detail:
			change.Change = BundleChangeChanged
		default:
			change.Change = BundleChangeUnchanged
		}
		if item.kind == BundleContentKindComponent && change.Change != BundleChangeUnchanged && change.Change != BundleChangeRemoved {
			change.PlanStepID = "deploy-" + item.name + "-plan"
			change.ApplyStepID = "deploy-" + item.name + "-apply"
		}
		if item.kind == BundleContentKindSandbox && change.Change == BundleChangeChanged {
			change.PlanStepID = "sandbox-plan"
			change.ApplyStepID = "sandbox-apply"
		}
		changes = append(changes, change)
	}
	return changes
}

type BundleTarget struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

// BundleVerification records the checks the runner performed while loading
// the bundle; the portal renders these as facts, not re-runnable checks.
type BundleVerification struct {
	BlobsVerified  bool `json:"blobs_verified"`
	EnvelopeParsed bool `json:"envelope_parsed"`
}

type BundleContent struct {
	Kind                string                   `json:"kind"`
	Name                string                   `json:"name"`
	Detail              string                   `json:"detail,omitempty"`
	Digest              string                   `json:"digest,omitempty"`
	ConfigDigest        string                   `json:"config_digest,omitempty"`
	Size                int64                    `json:"size,omitempty"`
	ComponentDefinition map[string]any           `json:"component_definition,omitempty"`
	ActionDefinition    *BundleActionDefinition  `json:"action_definition,omitempty"`
	RunbookDefinition   *BundleRunbookDefinition `json:"runbook_definition,omitempty"`
}

type BundleActionDefinition struct {
	TimeoutNanos          int64                 `json:"timeout_nanos,omitempty"`
	Role                  string                `json:"role,omitempty"`
	BreakGlassRoleARN     string                `json:"break_glass_role_arn,omitempty"`
	EnableKubeConfig      bool                  `json:"enable_kube_config,omitempty"`
	KubernetesContextName string                `json:"kubernetes_context_name,omitempty"`
	ComponentDependencies []string              `json:"component_dependencies,omitempty"`
	References            []string              `json:"references,omitempty"`
	Triggers              []BundleActionTrigger `json:"triggers,omitempty"`
	Steps                 []BundleActionStep    `json:"steps,omitempty"`
}

type BundleActionTrigger struct {
	Type          string `json:"type"`
	Index         int    `json:"index,omitempty"`
	CronSchedule  string `json:"cron_schedule,omitempty"`
	ComponentName string `json:"component_name,omitempty"`
}

type BundleActionStep struct {
	Name                 string            `json:"name"`
	Command              string            `json:"command,omitempty"`
	InlineContentsDigest string            `json:"inline_contents_digest,omitempty"`
	Source               *BundleSource     `json:"source,omitempty"`
	ArtifactDigest       string            `json:"artifact_digest,omitempty"`
	Index                int               `json:"index,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
}

type BundleRunbookDefinition struct {
	ReadmeDigest string               `json:"readme_digest,omitempty"`
	Inputs       []BundleRunbookInput `json:"inputs,omitempty"`
	Steps        []BundleRunbookStep  `json:"steps"`
}

type BundleRunbookStep struct {
	Kind                 string            `json:"kind"`
	Name                 string            `json:"name,omitempty"`
	Index                int               `json:"index,omitempty"`
	Reference            string            `json:"reference,omitempty"`
	Component            string            `json:"component,omitempty"`
	Role                 string            `json:"role,omitempty"`
	PlanOnly             bool              `json:"plan_only,omitempty"`
	DeployDependents     bool              `json:"deploy_dependents,omitempty"`
	TearDownDependents   bool              `json:"tear_down_dependents,omitempty"`
	SkipComponentDeploys bool              `json:"skip_component_deploys,omitempty"`
	Command              string            `json:"command,omitempty"`
	InlineContentsDigest string            `json:"inline_contents_digest,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
	TimeoutNanos         int64             `json:"timeout_nanos,omitempty"`
	TriggerName          string            `json:"trigger_name,omitempty"`
	EventTypes           []string          `json:"event_types,omitempty"`
	FiltersDigest        string            `json:"filters_digest,omitempty"`
}

type BundleRunbookInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Type        string `json:"type,omitempty"`
	Index       int    `json:"index,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

type BundleSource struct {
	Repository   string `json:"repository,omitempty"`
	RequestedRef string `json:"requested_ref,omitempty"`
	Commit       string `json:"commit,omitempty"`
	Directory    string `json:"directory,omitempty"`
	Version      string `json:"version,omitempty"`
	Digest       string `json:"digest,omitempty"`
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
